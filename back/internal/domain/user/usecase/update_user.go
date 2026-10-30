package usecase

import (
	"context"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
)

type UpdateUser struct {
	service service.Service
}

func NewUpdateUser(svc service.Service) *UpdateUser {
	return &UpdateUser{service: svc}
}

// Execute permite que o admin altere qualquer usuário e que o usuário comum
// altere apenas os próprios dados, sem mudar o próprio papel.
func (uc *UpdateUser) Execute(ctx context.Context, actor Actor, id uint, input model.UpdateUserInput) (*model.User, error) {
	if !actor.CanAccess(id) {
		return nil, usererrors.ErrForbidden
	}
	if input.Role != nil && !actor.IsAdmin() {
		return nil, usererrors.ErrForbidden
	}

	return uc.service.Update(ctx, id, input)
}
