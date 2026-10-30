package errors

import (
	stderrors "errors"
	"strings"
)

var (
	ErrNotFound           = stderrors.New("usuário não encontrado")
	ErrInvalidCredentials = stderrors.New("login ou senha inválidos")
	ErrLoginTaken         = stderrors.New("login já cadastrado")
	ErrEmailTaken         = stderrors.New("email já cadastrado")
	ErrForbidden          = stderrors.New("sem permissão para esta operação")
	ErrCannotDeleteSelf   = stderrors.New("não é possível remover o próprio usuário")
	ErrSessionExpired     = stderrors.New("sessão expirada, faça login novamente")
)

// ValidationError agrupa erros de validação por campo.
type ValidationError struct {
	Fields map[string]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{Fields: map[string]string{}}
}

func (e *ValidationError) Add(field, message string) {
	e.Fields[field] = message
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

// OrNil devolve o próprio erro quando há campos inválidos, ou nil.
// Evita o problema de interface não-nil com ponteiro nil.
func (e *ValidationError) OrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for field, msg := range e.Fields {
		parts = append(parts, field+": "+msg)
	}
	return "dados inválidos: " + strings.Join(parts, "; ")
}

func IsValidation(err error) (*ValidationError, bool) {
	var vErr *ValidationError
	if stderrors.As(err, &vErr) {
		return vErr, true
	}
	return nil, false
}
