package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
)

var (
	ctx   = context.Background()
	admin = model.Actor{ID: 1, Role: model.RoleAdmin}
	buddy = model.Actor{ID: 2, Role: model.RoleUser}
)

// fakeRepo é um repositório em memória, para exercitar o fluxo service -> use cases.
type fakeRepo struct {
	repository.Repository
	users map[uint]*model.User
	next  uint
}

func (r *fakeRepo) Create(_ context.Context, u *model.User) error {
	r.next++
	u.ID = r.next
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) Update(_ context.Context, u *model.User) error {
	r.users[u.ID] = u
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.users[id]; !ok {
		return usererrors.ErrNotFound
	}
	delete(r.users, id)
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	if u, ok := r.users[id]; ok {
		copied := *u
		return &copied, nil
	}
	return nil, usererrors.ErrNotFound
}

func (r *fakeRepo) FindByLogin(_ context.Context, login string) (*model.User, error) {
	for _, u := range r.users {
		if u.Login == login {
			return u, nil
		}
	}
	return nil, usererrors.ErrNotFound
}

func (r *fakeRepo) List(_ context.Context, offset, limit int) ([]model.User, int64, error) {
	return []model.User{{ID: uint(offset), Name: "page"}}, int64(limit), nil
}

func (r *fakeRepo) ExistsByLogin(_ context.Context, login string, exceptID uint) (bool, error) {
	for _, u := range r.users {
		if u.Login == login && u.ID != exceptID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepo) ExistsByEmail(_ context.Context, email string, exceptID uint) (bool, error) {
	for _, u := range r.users {
		if u.Email == email && u.ID != exceptID {
			return true, nil
		}
	}
	return false, nil
}

type fakeMailer struct{ sent chan mailer.Message }

func (m *fakeMailer) Send(msg mailer.Message) error {
	m.sent <- msg
	return nil
}

type fixture struct {
	svc    service.Service
	repo   *fakeRepo
	mail   *fakeMailer
	tokens *token.Manager
}

func newFixture() fixture {
	repo := &fakeRepo{users: map[uint]*model.User{}}
	mail := &fakeMailer{sent: make(chan mailer.Message, 1)}
	tokens := token.NewManager("test-secret", time.Hour, "test")
	return fixture{
		svc: service.New(service.Deps{
			Repo:           repo,
			PasswordSecret: "pepper",
			Tokens:         tokens,
			RefreshGrace:   time.Hour,
			Mailer:         mail,
		}),
		repo:   repo,
		mail:   mail,
		tokens: tokens,
	}
}

func validInput() model.CreateUserInput {
	return model.CreateUserInput{
		Login:    "  maria  ",
		Name:     " Maria ",
		Email:    " MARIA@Example.com ",
		Password: "12345678",
	}
}

func TestCreateRunsFullFlowAndSendsWelcomeEmail(t *testing.T) {
	f := newFixture()

	user, err := f.svc.Create(ctx, admin, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if user.Login != "maria" || user.Name != "Maria" || user.Email != "maria@example.com" {
		t.Errorf("campos não normalizados: %+v", user)
	}
	if user.Role != model.RoleUser {
		t.Errorf("papel padrão = %q", user.Role)
	}
	if user.Password == "12345678" || user.Password == "" {
		t.Errorf("senha não virou hash: %q", user.Password)
	}

	select {
	case msg := <-f.mail.sent:
		if msg.To[0] != "maria@example.com" {
			t.Errorf("destinatário = %v", msg.To)
		}
	case <-time.After(time.Second):
		t.Error("e-mail de boas-vindas não foi enviado")
	}
}

func TestCreateRejections(t *testing.T) {
	f := newFixture()

	if _, err := f.svc.Create(ctx, buddy, validInput()); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("não admin: %v", err)
	}
	if len(f.repo.users) != 0 {
		t.Error("não admin não deveria gravar")
	}

	invalid := model.CreateUserInput{Login: "x", Email: "invalido", Password: "curta", Role: "root"}
	_, err := f.svc.Create(ctx, admin, invalid)
	vErr, ok := usererrors.IsValidation(err)
	if !ok {
		t.Fatalf("esperava erro de validação: %v", err)
	}
	for _, field := range []string{"login", "name", "email", "password", "role"} {
		if _, ok := vErr.Fields[field]; !ok {
			t.Errorf("campo %q sem erro de validação", field)
		}
	}

	if _, err := f.svc.Create(ctx, admin, validInput()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	<-f.mail.sent
	if _, err := f.svc.Create(ctx, admin, validInput()); !errors.Is(err, usererrors.ErrLoginTaken) {
		t.Errorf("login duplicado: %v", err)
	}
}

func TestProvisionDoesNotRequireMailerOrActor(t *testing.T) {
	repo := &fakeRepo{users: map[uint]*model.User{}}
	// Mesma construção do seeder: sem mailer e sem token manager.
	svc := service.New(service.Deps{Repo: repo, PasswordSecret: "pepper"})

	input := validInput()
	input.Role = model.RoleAdmin
	user, err := svc.Provision(ctx, input)
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if user.Role != model.RoleAdmin || repo.users[user.ID] == nil {
		t.Errorf("usuário não gravado como admin: %+v", user)
	}
}

func TestLoginAndRefresh(t *testing.T) {
	f := newFixture()
	created, err := f.svc.Provision(ctx, validInput())
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	out, err := f.svc.Login(ctx, model.LoginInput{Login: " maria ", Password: "12345678"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	claims, err := f.tokens.Parse(out.Token)
	if err != nil || claims.UserID != created.ID {
		t.Errorf("token inválido: %v %+v", err, claims)
	}

	for name, input := range map[string]model.LoginInput{
		"senha errada":      {Login: "maria", Password: "errada123"},
		"login inexistente": {Login: "joao", Password: "12345678"},
	} {
		if _, err := f.svc.Login(ctx, input); !errors.Is(err, usererrors.ErrInvalidCredentials) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := f.svc.Login(ctx, model.LoginInput{Login: "  "}); err == nil {
		t.Error("login vazio deveria falhar na validação")
	} else if _, ok := usererrors.IsValidation(err); !ok {
		t.Errorf("login vazio: %v", err)
	}

	f.repo.users[created.ID].Role = model.RoleAdmin
	refreshed, err := f.svc.Refresh(ctx, out.Token)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if claims, _ := f.tokens.Parse(refreshed.Token); claims == nil || claims.Role != string(model.RoleAdmin) {
		t.Errorf("novo token sem o papel atualizado: %+v", claims)
	}

	if _, err := f.svc.Refresh(ctx, "token-invalido"); !errors.Is(err, usererrors.ErrSessionExpired) {
		t.Errorf("token inválido: %v", err)
	}
	delete(f.repo.users, created.ID)
	if _, err := f.svc.Refresh(ctx, out.Token); !errors.Is(err, usererrors.ErrSessionExpired) {
		t.Errorf("usuário removido: %v", err)
	}
}

func TestGetAndListAuthorization(t *testing.T) {
	f := newFixture()
	f.repo.users[2] = &model.User{ID: 2, Login: "buddy"}
	f.repo.users[3] = &model.User{ID: 3, Login: "other"}

	if _, err := f.svc.Get(ctx, buddy, 2); err != nil {
		t.Errorf("próprio usuário: %v", err)
	}
	if _, err := f.svc.Get(ctx, buddy, 3); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("outro usuário: %v", err)
	}
	if _, err := f.svc.Get(ctx, admin, 99); !errors.Is(err, usererrors.ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}

	if _, err := f.svc.List(ctx, buddy, model.ListUsersInput{}); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("listar sem ser admin: %v", err)
	}
	page, err := f.svc.List(ctx, admin, model.ListUsersInput{Page: 3, PerPage: 1000})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.Page != 3 || page.PerPage != 100 {
		t.Errorf("paginação não normalizada: %+v", page)
	}
	if items := page.Items.([]model.User); items[0].ID != 200 {
		t.Errorf("offset = %d", items[0].ID)
	}
}

func TestUpdate(t *testing.T) {
	f := newFixture()
	f.repo.users[2] = &model.User{ID: 2, Login: "buddy", Name: "Buddy", Email: "buddy@example.com", Password: "old", Role: model.RoleUser}
	f.repo.users[3] = &model.User{ID: 3, Login: "other", Name: "Other", Email: "other@example.com", Role: model.RoleUser}

	name, taken, role, short, pass := " Novo ", "other", model.RoleAdmin, "curta", "nova-senha"

	if _, err := f.svc.Update(ctx, buddy, 3, model.UpdateUserInput{Name: &name}); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("alterar outro usuário: %v", err)
	}
	if _, err := f.svc.Update(ctx, buddy, 2, model.UpdateUserInput{Role: &role}); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("alterar o próprio papel: %v", err)
	}
	if _, err := f.svc.Update(ctx, buddy, 2, model.UpdateUserInput{Password: &short}); err == nil {
		t.Error("senha curta deveria falhar")
	} else if _, ok := usererrors.IsValidation(err); !ok {
		t.Errorf("senha curta: %v", err)
	}
	if _, err := f.svc.Update(ctx, buddy, 2, model.UpdateUserInput{Login: &taken}); !errors.Is(err, usererrors.ErrLoginTaken) {
		t.Errorf("login em uso: %v", err)
	}
	if f.repo.users[2].Login != "buddy" {
		t.Error("atualização rejeitada não deveria alterar o usuário gravado")
	}

	user, err := f.svc.Update(ctx, buddy, 2, model.UpdateUserInput{Name: &name, Password: &pass})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if user.Name != "Novo" || user.Password == "old" || user.Password == pass {
		t.Errorf("alterações não aplicadas: %+v", user)
	}

	if user, err = f.svc.Update(ctx, admin, 3, model.UpdateUserInput{Role: &role}); err != nil || user.Role != role {
		t.Errorf("admin alterando papel: %v %+v", err, user)
	}
}

func TestDelete(t *testing.T) {
	f := newFixture()
	f.repo.users[1] = &model.User{ID: 1}
	f.repo.users[2] = &model.User{ID: 2}

	if err := f.svc.Delete(ctx, buddy, 1); !errors.Is(err, usererrors.ErrForbidden) {
		t.Errorf("sem ser admin: %v", err)
	}
	if err := f.svc.Delete(ctx, admin, 1); !errors.Is(err, usererrors.ErrCannotDeleteSelf) {
		t.Errorf("a si mesmo: %v", err)
	}
	if err := f.svc.Delete(ctx, admin, 2); err != nil {
		t.Errorf("Delete: %v", err)
	}
	if err := f.svc.Delete(ctx, admin, 2); !errors.Is(err, usererrors.ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}
