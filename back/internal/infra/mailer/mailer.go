package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
)

type Message struct {
	To      []string
	Subject string
	Body    string // texto puro
}

type Mailer interface {
	Send(msg Message) error
}

// New retorna um mailer SMTP quando MAIL_ENABLED=true; caso contrário
// retorna um mailer que apenas registra o e-mail no log (útil em dev).
func New(cfg config.MailConfig) Mailer {
	if !cfg.Enabled {
		return logMailer{}
	}
	return &smtpMailer{cfg: cfg}
}

type smtpMailer struct {
	cfg config.MailConfig
}

func (m *smtpMailer) Send(msg Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("mailer: nenhum destinatário informado")
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	if err := smtp.SendMail(addr, auth, m.cfg.From, msg.To, m.build(msg)); err != nil {
		return fmt.Errorf("mailer: erro ao enviar e-mail: %w", err)
	}
	return nil
}

func (m *smtpMailer) build(msg Message) []byte {
	var b strings.Builder
	b.WriteString("From: " + m.cfg.From + "\r\n")
	b.WriteString("To: " + strings.Join(msg.To, ", ") + "\r\n")
	b.WriteString("Subject: " + msg.Subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)
	return []byte(b.String())
}

type logMailer struct{}

func (logMailer) Send(msg Message) error {
	log.Printf("mailer (desabilitado): para=%v assunto=%q\n%s", msg.To, msg.Subject, msg.Body)
	return nil
}
