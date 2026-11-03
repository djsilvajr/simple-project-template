package usecase

import (
	"fmt"
	"log"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
)

// SendWelcomeEmail envia o e-mail de boas-vindas. A falha é apenas registrada
// no log: não deve desfazer o cadastro.
func SendWelcomeEmail(m mailer.Mailer, user model.User) {
	err := m.Send(mailer.Message{
		To:      []string{user.Email},
		Subject: "Bem-vindo(a)!",
		Body:    fmt.Sprintf("Olá, %s!\n\nSua conta foi criada. Seu login é: %s\n", user.Name, user.Login),
	})
	if err != nil {
		log.Printf("usecase: falha ao enviar e-mail de boas-vindas para usuário %d: %v", user.ID, err)
	}
}
