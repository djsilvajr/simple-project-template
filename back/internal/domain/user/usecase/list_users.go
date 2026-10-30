package usecase

import (
	"context"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

type ListUsers struct {
	service service.Service
}

func NewListUsers(svc service.Service) *ListUsers {
	return &ListUsers{service: svc}
}

func (uc *ListUsers) Execute(ctx context.Context, actor Actor, input model.ListUsersInput) (*response.Paginated, error) {
	if !actor.IsAdmin() {
		return nil, usererrors.ErrForbidden
	}

	users, total, page, err := uc.service.List(ctx, input)
	if err != nil {
		return nil, err
	}

	return &response.Paginated{
		Items:   users,
		Total:   total,
		Page:    page.Page,
		PerPage: page.PerPage,
	}, nil
}
