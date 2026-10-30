package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	defaultPerPage    = 20
	maxPerPage        = 100
)

var loginPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,100}$`)

// Service concentra as regras de domínio do usuário (validação, unicidade,
// hash de senha). É reutilizado pelos use cases e pelo seeder.
type Service interface {
	Create(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	Update(ctx context.Context, id uint, input model.UpdateUserInput) (*model.User, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	List(ctx context.Context, input model.ListUsersInput) ([]model.User, int64, model.ListUsersInput, error)
	Authenticate(ctx context.Context, login, password string) (*model.User, error)
}

type service struct {
	repo           repository.Repository
	passwordSecret []byte
}

func New(repo repository.Repository, passwordSecret string) Service {
	return &service{
		repo:           repo,
		passwordSecret: []byte(passwordSecret),
	}
}

func (s *service) Create(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	input.Login = strings.TrimSpace(input.Login)
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Role == "" {
		input.Role = model.RoleUser
	}

	vErr := usererrors.NewValidationError()
	validateLogin(vErr, input.Login)
	validateName(vErr, input.Name)
	validateEmail(vErr, input.Email)
	validatePassword(vErr, input.Password)
	validateRole(vErr, input.Role)
	if err := vErr.OrNil(); err != nil {
		return nil, err
	}

	if err := s.ensureUnique(ctx, input.Login, input.Email, 0); err != nil {
		return nil, err
	}

	hash, err := s.hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Login:    input.Login,
		Name:     input.Name,
		Email:    input.Email,
		Password: hash,
		Role:     input.Role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, translateDBError(err)
	}

	return user, nil
}

func (s *service) Update(ctx context.Context, id uint, input model.UpdateUserInput) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	vErr := usererrors.NewValidationError()

	if input.Login != nil {
		user.Login = strings.TrimSpace(*input.Login)
		validateLogin(vErr, user.Login)
	}
	if input.Name != nil {
		user.Name = strings.TrimSpace(*input.Name)
		validateName(vErr, user.Name)
	}
	if input.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*input.Email))
		validateEmail(vErr, user.Email)
	}
	if input.Role != nil {
		user.Role = *input.Role
		validateRole(vErr, user.Role)
	}
	if input.Password != nil {
		validatePassword(vErr, *input.Password)
	}
	if err := vErr.OrNil(); err != nil {
		return nil, err
	}

	if err := s.ensureUnique(ctx, user.Login, user.Email, user.ID); err != nil {
		return nil, err
	}

	if input.Password != nil {
		if user.Password, err = s.hashPassword(*input.Password); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, translateDBError(err)
	}

	return user, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id uint) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

// List normaliza a paginação e devolve os valores efetivamente usados.
func (s *service) List(ctx context.Context, input model.ListUsersInput) ([]model.User, int64, model.ListUsersInput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 {
		input.PerPage = defaultPerPage
	}
	if input.PerPage > maxPerPage {
		input.PerPage = maxPerPage
	}

	users, total, err := s.repo.List(ctx, (input.Page-1)*input.PerPage, input.PerPage)
	return users, total, input, err
}

// Authenticate valida login e senha. Retorna sempre ErrInvalidCredentials
// (sem distinguir usuário inexistente de senha errada).
func (s *service) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	login = strings.TrimSpace(login)

	vErr := usererrors.NewValidationError()
	if login == "" {
		vErr.Add("login", "obrigatório")
	}
	if password == "" {
		vErr.Add("password", "obrigatório")
	}
	if err := vErr.OrNil(); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByLogin(ctx, login)
	if errors.Is(err, usererrors.ErrNotFound) {
		// Executa um bcrypt mesmo assim para não vazar, pelo tempo de
		// resposta, se o login existe.
		_ = bcrypt.CompareHashAndPassword(dummyHash, s.pepper(password))
		return nil, usererrors.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !s.checkPassword(user.Password, password) {
		return nil, usererrors.ErrInvalidCredentials
	}

	return user, nil
}

// --- senha -----------------------------------------------------------------

// dummyHash é usado apenas para igualar o tempo de resposta quando o login não existe.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// pepper aplica HMAC-SHA256 com PASSWORD_SECRET antes do bcrypt. Isso adiciona
// o "pepper" e mantém a entrada abaixo do limite de 72 bytes do bcrypt.
func (s *service) pepper(password string) []byte {
	mac := hmac.New(sha256.New, s.passwordSecret)
	mac.Write([]byte(password))
	return []byte(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(s.pepper(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *service) checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), s.pepper(password)) == nil
}

// --- unicidade / validação -------------------------------------------------

func (s *service) ensureUnique(ctx context.Context, login, email string, exceptID uint) error {
	exists, err := s.repo.ExistsByLogin(ctx, login, exceptID)
	if err != nil {
		return err
	}
	if exists {
		return usererrors.ErrLoginTaken
	}

	exists, err = s.repo.ExistsByEmail(ctx, email, exceptID)
	if err != nil {
		return err
	}
	if exists {
		return usererrors.ErrEmailTaken
	}

	return nil
}

// translateDBError cobre a corrida entre ensureUnique e o INSERT/UPDATE:
// o único índice único da tabela é o de email.
func translateDBError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return usererrors.ErrEmailTaken
	}
	return err
}

func validateLogin(vErr *usererrors.ValidationError, login string) {
	if !loginPattern.MatchString(login) {
		vErr.Add("login", "deve ter de 3 a 100 caracteres (letras, números, '.', '_' ou '-')")
	}
}

func validateName(vErr *usererrors.ValidationError, name string) {
	if name == "" {
		vErr.Add("name", "obrigatório")
	} else if utf8.RuneCountInString(name) > 100 {
		vErr.Add("name", "deve ter no máximo 100 caracteres")
	}
}

// Email é obrigatório: a coluna tem índice único e o tipo string grava ""
// quando vazio, o que faria dois usuários sem email colidirem.
func validateEmail(vErr *usererrors.ValidationError, email string) {
	if email == "" {
		vErr.Add("email", "obrigatório")
		return
	}
	if utf8.RuneCountInString(email) > 150 {
		vErr.Add("email", "deve ter no máximo 150 caracteres")
		return
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		vErr.Add("email", "formato inválido")
	}
}

func validatePassword(vErr *usererrors.ValidationError, password string) {
	n := utf8.RuneCountInString(password)
	if n < minPasswordLength {
		vErr.Add("password", "deve ter no mínimo 8 caracteres")
	} else if n > maxPasswordLength {
		vErr.Add("password", "deve ter no máximo 128 caracteres")
	}
}

func validateRole(vErr *usererrors.ValidationError, role model.Role) {
	if !role.IsValid() {
		vErr.Add("role", "deve ser 'admin' ou 'user'")
	}
}
