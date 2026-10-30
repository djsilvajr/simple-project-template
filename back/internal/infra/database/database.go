package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
)

const (
	connectAttempts = 10
	connectInterval = 3 * time.Second
)

// Connect abre a conexão com o MySQL, configura o pool e tenta novamente
// algumas vezes caso o banco ainda esteja subindo.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.App.Debug {
		logLevel = logger.Info
	}

	var (
		db  *gorm.DB
		err error
	)

	for attempt := 1; attempt <= connectAttempts; attempt++ {
		db, err = gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{
			Logger:         newLogger(logLevel),
			TranslateError: true,
		})
		if err == nil {
			break
		}

		log.Printf("database: tentativa %d/%d falhou: %v", attempt, connectAttempts, err)
		time.Sleep(connectInterval)
	}
	if err != nil {
		return nil, fmt.Errorf("database: não foi possível conectar: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: erro ao obter sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping falhou: %w", err)
	}

	return db, nil
}

// newLogger configura o log do GORM. "record not found" não é tratado como
// erro, já que o repositório converte isso em ErrNotFound.
func newLogger(level logger.LogLevel) logger.Interface {
	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
}

// Close fecha o pool de conexões.
func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("database: erro ao fechar conexão: %v", err)
	}
}
