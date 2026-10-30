package usecase

import (
	"context"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
)

type GetUser struct {
	service service.Service
}

func NewGetUser(svc service.Service) *GetUser {
	return &GetUser{service: svc}
}

func (uc *GetUser) Execute(ctx context.Context, actor Actor, id uint) (*model.User, error) {
	if !actor.CanAccess(id) {
		return nil, usererrors.ErrForbidden
	}
	return uc.service.GetByID(ctx, id)
}
