package usecase

import "github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"

// Actor é o usuário autenticado que está executando a ação.
type Actor struct {
	ID   uint
	Role model.Role
}

func (a Actor) IsAdmin() bool {
	return a.Role == model.RoleAdmin
}

// CanAccess indica se o ator pode ver/alterar o usuário informado:
// admin acessa qualquer usuário, user apenas a si mesmo.
func (a Actor) CanAccess(userID uint) bool {
	return a.IsAdmin() || a.ID == userID
}
