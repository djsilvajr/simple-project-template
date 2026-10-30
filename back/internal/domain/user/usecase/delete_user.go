package usecase

import (
	"context"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
)

type DeleteUser struct {
	service service.Service
}

func NewDeleteUser(svc service.Service) *DeleteUser {
	return &DeleteUser{service: svc}
}

func (uc *DeleteUser) Execute(ctx context.Context, actor Actor, id uint) error {
	if !actor.IsAdmin() {
		return usererrors.ErrForbidden
	}
	if actor.ID == id {
		return usererrors.ErrCannotDeleteSelf
	}

	return uc.service.Delete(ctx, id)
}
