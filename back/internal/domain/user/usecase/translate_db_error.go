package usecase

import (
	"errors"

	"gorm.io/gorm"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
)

// translateDBError cobre a corrida entre EnsureUserUnique e o INSERT/UPDATE:
// o único índice único da tabela é o de email.
func translateDBError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return usererrors.ErrEmailTaken
	}
	return err
}
