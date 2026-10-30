package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Mail      MailConfig
	Seed      SeedConfig
}

type AppConfig struct {
	Name           string
	Env            string
	Port           int
	Debug          bool
	LogRequestBody bool
	// TrustProxy faz o IP do cliente ser lido de X-Real-IP / X-Forwarded-For.
	// Só ative quando a API estiver atrás de um proxy confiável (nginx).
	TrustProxy bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	// JWTRefreshGrace é quanto tempo depois de expirar um token ainda pode ser renovado.
	JWTRefreshGrace time.Duration
	PasswordSecret  string
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	Burst             int
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type MailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SeedConfig struct {
	AdminLogin    string
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

// Load carrega o arquivo .env (se existir) e monta o Config a partir das
// variáveis de ambiente. Variáveis já definidas no ambiente têm prioridade
// sobre as do arquivo.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("config: erro ao ler .env: %w", err)
	} else if err != nil {
		log.Println("config: arquivo .env não encontrado, usando apenas variáveis de ambiente")
	}

	cfg := &Config{
		App: AppConfig{
			Name:           getEnvStr("APP_NAME", "api"),
			Env:            getEnvStr("APP_ENV", "local"),
			Port:           getEnvInt("APP_PORT", 8080),
			Debug:          getEnvBool("APP_DEBUG", false),
			LogRequestBody: getEnvBool("LOG_REQUEST_BODY", false),
			TrustProxy:     getEnvBool("TRUST_PROXY", false),
		},
		Database: DatabaseConfig{
			Host:            getEnvStr("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 3306),
			Name:            getEnvStr("DB_DATABASE", ""),
			User:            getEnvStr("DB_USERNAME", ""),
			Password:        getEnvStr("DB_PASSWORD", ""),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
		},
		Auth: AuthConfig{
			JWTSecret:       getEnvStr("JWT_SECRET", ""),
			JWTExpiration:   time.Duration(getEnvInt("JWT_EXPIRATION_MINUTES", 60)) * time.Minute,
			JWTRefreshGrace: time.Duration(getEnvInt("JWT_REFRESH_GRACE_MINUTES", 30)) * time.Minute,
			PasswordSecret:  getEnvStr("PASSWORD_SECRET", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: getEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 30),
			Burst:             getEnvInt("RATE_LIMIT_BURST", 30),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitList(getEnvStr("CORS_ALLOWED_ORIGINS", "*")),
			AllowedMethods:   splitList(getEnvStr("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")),
			AllowedHeaders:   splitList(getEnvStr("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvInt("CORS_MAX_AGE", 600),
		},
		Mail: MailConfig{
			Enabled:  getEnvBool("MAIL_ENABLED", false),
			Host:     getEnvStr("MAIL_HOST", ""),
			Port:     getEnvInt("MAIL_PORT", 587),
			Username: getEnvStr("MAIL_USERNAME", ""),
			Password: getEnvStr("MAIL_PASSWORD", ""),
			From:     getEnvStr("MAIL_FROM", "no-reply@example.com"),
		},
		Seed: SeedConfig{
			AdminLogin:    getEnvStr("SEED_ADMIN_LOGIN", "admin"),
			AdminName:     getEnvStr("SEED_ADMIN_NAME", "Administrador"),
			AdminEmail:    getEnvStr("SEED_ADMIN_EMAIL", "admin@example.com"),
			AdminPassword: getEnvStr("SEED_ADMIN_PASSWORD", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN monta a string de conexão do MySQL.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

func (c *Config) validate() error {
	var missing []string

	if c.Database.Name == "" {
		missing = append(missing, "DB_DATABASE")
	}
	if c.Database.User == "" {
		missing = append(missing, "DB_USERNAME")
	}
	if c.Auth.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.Auth.PasswordSecret == "" {
		missing = append(missing, "PASSWORD_SECRET")
	}
	if c.Mail.Enabled && c.Mail.Host == "" {
		missing = append(missing, "MAIL_HOST")
	}

	if len(missing) > 0 {
		return fmt.Errorf("config: variáveis obrigatórias ausentes: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		log.Printf("config: valor inválido para %s=%q, usando padrão %v", key, value, fallback)
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		log.Printf("config: valor inválido para %s=%q, usando padrão %d", key, value, fallback)
		return fallback
	}
	return parsed
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
