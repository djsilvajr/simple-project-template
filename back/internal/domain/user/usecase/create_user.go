package usecase

import (
	"context"
	"fmt"
	"log"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
)

type CreateUser struct {
	service service.Service
	mailer  mailer.Mailer
}

func NewCreateUser(svc service.Service, m mailer.Mailer) *CreateUser {
	return &CreateUser{service: svc, mailer: m}
}

// Execute cria o usuário (apenas admin) e envia um e-mail de boas-vindas
// em segundo plano. Falha no envio não desfaz o cadastro.
func (uc *CreateUser) Execute(ctx context.Context, actor Actor, input model.CreateUserInput) (*model.User, error) {
	if !actor.IsAdmin() {
		return nil, usererrors.ErrForbidden
	}

	user, err := uc.service.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	go uc.sendWelcome(*user)

	return user, nil
}

func (uc *CreateUser) sendWelcome(user model.User) {
	err := uc.mailer.Send(mailer.Message{
		To:      []string{user.Email},
		Subject: "Bem-vindo(a)!",
		Body:    fmt.Sprintf("Olá, %s!\n\nSua conta foi criada. Seu login é: %s\n", user.Name, user.Login),
	})
	if err != nil {
		log.Printf("usecase: falha ao enviar e-mail de boas-vindas para usuário %d: %v", user.ID, err)
	}
}
