package service

import (
	"context"
	"time"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/usecase"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

// Service orquestra as operações do domínio: cada método define o fluxo de uma
// operação e delega cada passo a um use case. É a única camada chamada pelo
// handler e não contém regra de negócio própria.
type Service interface {
	Login(ctx context.Context, input model.LoginInput) (*model.LoginOutput, error)
	Refresh(ctx context.Context, rawToken string) (*model.LoginOutput, error)
	Create(ctx context.Context, actor model.Actor, input model.CreateUserInput) (*model.User, error)
	Provision(ctx context.Context, input model.CreateUserInput) (*model.User, error)
	Get(ctx context.Context, actor model.Actor, id uint) (*model.User, error)
	List(ctx context.Context, actor model.Actor, input model.ListUsersInput) (*response.Paginated, error)
	Update(ctx context.Context, actor model.Actor, id uint, input model.UpdateUserInput) (*model.User, error)
	Delete(ctx context.Context, actor model.Actor, id uint) error
}

// Deps reúne as dependências do service. Provision usa apenas Repo e
// PasswordSecret, então o seeder não precisa informar as demais.
type Deps struct {
	Repo           repository.Repository
	PasswordSecret string
	Tokens         *token.Manager
	RefreshGrace   time.Duration
	Mailer         mailer.Mailer
}

type service struct {
	repo           repository.Repository
	passwordSecret []byte
	tokens         *token.Manager
	refreshGrace   time.Duration
	mailer         mailer.Mailer
}

func New(deps Deps) Service {
	return &service{
		repo:           deps.Repo,
		passwordSecret: []byte(deps.PasswordSecret),
		tokens:         deps.Tokens,
		refreshGrace:   deps.RefreshGrace,
		mailer:         deps.Mailer,
	}
}

func (s *service) Login(ctx context.Context, input model.LoginInput) (*model.LoginOutput, error) {
	input = usecase.NormalizeLoginInput(input)

	if err := usecase.ValidateLoginInput(input); err != nil {
		return nil, err
	}

	user, err := usecase.AuthenticateUser(ctx, s.repo, s.passwordSecret, input)
	if err != nil {
		return nil, err
	}

	return usecase.IssueToken(s.tokens, user)
}

// Refresh troca um token válido, ou expirado dentro da janela de renovação, por
// um novo. O usuário é recarregado do banco: se foi removido a renovação falha,
// e se o papel mudou o novo token já sai com o papel atualizado.
func (s *service) Refresh(ctx context.Context, rawToken string) (*model.LoginOutput, error) {
	claims, err := usecase.ParseRefreshToken(s.tokens, rawToken, s.refreshGrace)
	if err != nil {
		return nil, err
	}

	user, err := usecase.FindSessionUser(ctx, s.repo, claims.UserID)
	if err != nil {
		return nil, err
	}

	return usecase.IssueToken(s.tokens, user)
}

// Create cadastra o usuário (apenas admin) e envia o e-mail de boas-vindas em
// segundo plano.
func (s *service) Create(ctx context.Context, actor model.Actor, input model.CreateUserInput) (*model.User, error) {
	if err := usecase.AuthorizeAdmin(actor); err != nil {
		return nil, err
	}

	user, err := s.Provision(ctx, input)
	if err != nil {
		return nil, err
	}

	go usecase.SendWelcomeEmail(s.mailer, *user)

	return user, nil
}

// Provision cadastra um usuário sem ator e sem notificação. Existe para fluxos
// internos, como o seeder. As rotas da API devem usar Create.
func (s *service) Provision(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	input = usecase.NormalizeCreateUser(input)

	if err := usecase.ValidateCreateUser(input); err != nil {
		return nil, err
	}
	if err := usecase.EnsureUserUnique(ctx, s.repo, input.Login, input.Email, 0); err != nil {
		return nil, err
	}

	hash, err := usecase.HashPassword(input.Password, s.passwordSecret)
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
	if err := usecase.AddUser(ctx, s.repo, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Get(ctx context.Context, actor model.Actor, id uint) (*model.User, error) {
	if err := usecase.AuthorizeUserAccess(actor, id); err != nil {
		return nil, err
	}

	return usecase.FindUser(ctx, s.repo, id)
}

func (s *service) List(ctx context.Context, actor model.Actor, input model.ListUsersInput) (*response.Paginated, error) {
	if err := usecase.AuthorizeAdmin(actor); err != nil {
		return nil, err
	}

	input = usecase.NormalizePagination(input)

	users, total, err := usecase.ListUsers(ctx, s.repo, input)
	if err != nil {
		return nil, err
	}

	return &response.Paginated{
		Items:   users,
		Total:   total,
		Page:    input.Page,
		PerPage: input.PerPage,
	}, nil
}

// Update permite que o admin altere qualquer usuário e que o usuário comum
// altere apenas os próprios dados, sem mudar o próprio papel.
func (s *service) Update(ctx context.Context, actor model.Actor, id uint, input model.UpdateUserInput) (*model.User, error) {
	if err := usecase.AuthorizeUserAccess(actor, id); err != nil {
		return nil, err
	}
	if err := usecase.AuthorizeRoleChange(actor, input.Role); err != nil {
		return nil, err
	}

	user, err := usecase.FindUser(ctx, s.repo, id)
	if err != nil {
		return nil, err
	}

	usecase.ApplyUserChanges(user, input)

	if err := usecase.ValidateUpdateUser(user, input); err != nil {
		return nil, err
	}
	if err := usecase.EnsureUserUnique(ctx, s.repo, user.Login, user.Email, user.ID); err != nil {
		return nil, err
	}

	if input.Password != nil {
		if user.Password, err = usecase.HashPassword(*input.Password, s.passwordSecret); err != nil {
			return nil, err
		}
	}

	if err := usecase.SaveUser(ctx, s.repo, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Delete(ctx context.Context, actor model.Actor, id uint) error {
	if err := usecase.AuthorizeAdmin(actor); err != nil {
		return err
	}
	if err := usecase.EnsureNotSelf(actor, id); err != nil {
		return err
	}

	return usecase.RemoveUser(ctx, s.repo, id)
}
