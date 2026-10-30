package migration

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	usermodel "github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// models lista todas as entidades que devem ter tabela no banco.
// Ao criar um novo domínio, adicione o model aqui.
var models = []any{
	&usermodel.User{},
}

// Migrate cria/atualiza as tabelas a partir dos models (GORM AutoMigrate).
// AutoMigrate cria tabelas, colunas e índices que faltam, mas não remove
// colunas nem altera dados existentes.
func Migrate(db *gorm.DB) error {
	for _, m := range models {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(m); err != nil {
			return fmt.Errorf("erro ao ler o model %T: %w", m, err)
		}
		table := stmt.Schema.Table
		existed := db.Migrator().HasTable(m)

		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("erro ao migrar %T (tabela %s): %w", m, table, err)
		}

		if existed {
			log.Printf("migration: tabela %q já existia — colunas/índices verificados", table)
		} else {
			log.Printf("migration: tabela %q criada", table)
		}
	}
	return nil
}
