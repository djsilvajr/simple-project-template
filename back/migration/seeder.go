package migration

import (
	"context"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
)

// Seeder popula dados iniciais. Deve ser idempotente (pode rodar várias vezes).
type Seeder func(ctx context.Context, db *gorm.DB, cfg *config.Config) error

// seeders lista os seeders na ordem de execução.
var seeders = []struct {
	name string
	run  Seeder
}{
	{"admin user", seedAdminUser},
}

func Seed(db *gorm.DB, cfg *config.Config) error {
	ctx := context.Background()
	for _, s := range seeders {
		if err := s.run(ctx, db, cfg); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	return nil
}

// seedAdminUser cria o usuário admin inicial usando o mesmo service da API,
// garantindo as mesmas validações e o mesmo hash de senha.
func seedAdminUser(ctx context.Context, db *gorm.DB, cfg *config.Config) error {
	if cfg.Seed.AdminPassword == "" {
		log.Println("seeder: SEED_ADMIN_PASSWORD vazio, admin não será criado")
		return nil
	}

	repo := repository.New(db)

	_, err := repo.FindByLogin(ctx, cfg.Seed.AdminLogin)
	if err == nil {
		log.Printf("seeder: usuário %q já existe, ignorando", cfg.Seed.AdminLogin)
		return nil
	}
	if !errors.Is(err, usererrors.ErrNotFound) {
		return err
	}

	svc := service.New(service.Deps{Repo: repo, PasswordSecret: cfg.Auth.PasswordSecret})
	user, err := svc.Provision(ctx, model.CreateUserInput{
		Login:    cfg.Seed.AdminLogin,
		Name:     cfg.Seed.AdminName,
		Email:    cfg.Seed.AdminEmail,
		Password: cfg.Seed.AdminPassword,
		Role:     model.RoleAdmin,
	})
	if err != nil {
		return err
	}

	log.Printf("seeder: usuário admin %q criado (id=%d)", user.Login, user.ID)
	return nil
}
