package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]model.User, int64, error)
	ExistsByLogin(ctx context.Context, login string, exceptID uint) (bool, error)
	ExistsByEmail(ctx context.Context, email string, exceptID uint) (bool, error)
}

type gormRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete faz soft delete (preenche deleted_at).
func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return usererrors.ErrNotFound
	}
	return nil
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	return r.first(ctx, "id = ?", id)
}

func (r *gormRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	return r.first(ctx, "login = ?", login)
}

func (r *gormRepository) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var (
		users []model.User
		total int64
	)

	db := r.db.WithContext(ctx)

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("id ASC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *gormRepository) ExistsByLogin(ctx context.Context, login string, exceptID uint) (bool, error) {
	return r.exists(ctx, "login = ?", login, exceptID, false)
}

// ExistsByEmail usa Unscoped porque o índice único de email também vale
// para registros com soft delete.
func (r *gormRepository) ExistsByEmail(ctx context.Context, email string, exceptID uint) (bool, error) {
	return r.exists(ctx, "email = ?", email, exceptID, true)
}

func (r *gormRepository) first(ctx context.Context, query string, args ...any) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).Where(query, args...).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, usererrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *gormRepository) exists(ctx context.Context, query string, value any, exceptID uint, unscoped bool) (bool, error) {
	var count int64

	db := r.db.WithContext(ctx)
	if unscoped {
		db = db.Unscoped()
	}

	q := db.Model(&model.User{}).Where(query, value)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}

	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
