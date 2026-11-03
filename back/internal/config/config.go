package config

import (
	"fmt"
	"log"
	"math"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretBytes = 32

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
// sobre as do arquivo. Retorna erro, listando todos os problemas, quando algum
// valor é inválido: nesse caso a aplicação não deve subir.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("config: erro ao ler .env: %w", err)
	} else if err != nil {
		log.Println("config: arquivo .env não encontrado, usando apenas variáveis de ambiente")
	}

	env := &envReader{}

	cfg := &Config{
		App: AppConfig{
			Name:           env.str("APP_NAME", "api"),
			Env:            env.str("APP_ENV", "local"),
			Port:           env.integer("APP_PORT", 8080),
			Debug:          env.boolean("APP_DEBUG", false),
			LogRequestBody: env.boolean("LOG_REQUEST_BODY", false),
			TrustProxy:     env.boolean("TRUST_PROXY", false),
		},
		Database: DatabaseConfig{
			Host:            env.str("DB_HOST", "localhost"),
			Port:            env.integer("DB_PORT", 3306),
			Name:            env.str("DB_DATABASE", ""),
			User:            env.str("DB_USERNAME", ""),
			Password:        env.str("DB_PASSWORD", ""),
			MaxOpenConns:    env.integer("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    env.integer("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: env.minutes("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		},
		Auth: AuthConfig{
			JWTSecret:       env.str("JWT_SECRET", ""),
			JWTExpiration:   env.minutes("JWT_EXPIRATION_MINUTES", 60),
			JWTRefreshGrace: env.minutes("JWT_REFRESH_GRACE_MINUTES", 30),
			PasswordSecret:  env.str("PASSWORD_SECRET", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled:           env.boolean("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: env.integer("RATE_LIMIT_REQUESTS_PER_SECOND", 30),
			Burst:             env.integer("RATE_LIMIT_BURST", 30),
		},
		CORS: CORSConfig{
			AllowedOrigins:   env.list("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods:   env.list("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
			AllowedHeaders:   env.list("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
			AllowCredentials: env.boolean("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           env.integer("CORS_MAX_AGE", 600),
		},
		Mail: MailConfig{
			Enabled:  env.boolean("MAIL_ENABLED", false),
			Host:     env.str("MAIL_HOST", ""),
			Port:     env.integer("MAIL_PORT", 587),
			Username: env.str("MAIL_USERNAME", ""),
			Password: env.str("MAIL_PASSWORD", ""),
			From:     env.str("MAIL_FROM", "no-reply@example.com"),
		},
		Seed: SeedConfig{
			AdminLogin:    env.str("SEED_ADMIN_LOGIN", "admin"),
			AdminName:     env.str("SEED_ADMIN_NAME", "Administrador"),
			AdminEmail:    env.str("SEED_ADMIN_EMAIL", "admin@example.com"),
			AdminPassword: env.str("SEED_ADMIN_PASSWORD", ""),
		},
	}

	problems := append(env.problems, cfg.validate()...)
	if len(problems) > 0 {
		return nil, fmt.Errorf("config: env inválido:\n  - %s", strings.Join(problems, "\n  - "))
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

// validate confere as regras de cada valor já tipado e devolve todos os
// problemas encontrados, sem incluir valores de campos sensíveis.
func (c *Config) validate() []string {
	v := &validator{}

	v.required("APP_NAME", c.App.Name)
	v.required("APP_ENV", c.App.Env)
	v.port("APP_PORT", c.App.Port)

	v.required("DB_HOST", c.Database.Host)
	v.port("DB_PORT", c.Database.Port)
	v.required("DB_DATABASE", c.Database.Name)
	v.required("DB_USERNAME", c.Database.User)
	v.atLeast("DB_MAX_OPEN_CONNS", c.Database.MaxOpenConns, 1)
	v.atLeast("DB_MAX_IDLE_CONNS", c.Database.MaxIdleConns, 0)
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		v.addf("DB_MAX_IDLE_CONNS não pode ser maior que DB_MAX_OPEN_CONNS")
	}
	v.notNegative("DB_CONN_MAX_LIFETIME_MINUTES", c.Database.ConnMaxLifetime)

	// O HS256 exige uma chave de pelo menos 256 bits (RFC 7518, seção 3.2).
	if len(c.Auth.JWTSecret) < minJWTSecretBytes {
		v.addf("JWT_SECRET deve ter no mínimo %d bytes", minJWTSecretBytes)
	}
	v.required("PASSWORD_SECRET", c.Auth.PasswordSecret)
	if c.Auth.JWTExpiration <= 0 {
		v.addf("JWT_EXPIRATION_MINUTES deve ser maior que zero")
	}
	v.notNegative("JWT_REFRESH_GRACE_MINUTES", c.Auth.JWTRefreshGrace)

	if c.RateLimit.Enabled {
		v.atLeast("RATE_LIMIT_REQUESTS_PER_SECOND", c.RateLimit.RequestsPerSecond, 1)
		v.atLeast("RATE_LIMIT_BURST", c.RateLimit.Burst, 1)
	}

	for _, origin := range c.CORS.AllowedOrigins {
		if origin != "*" && !isOrigin(origin) {
			v.addf("CORS_ALLOWED_ORIGINS tem a origem inválida %q (use \"*\" ou http(s)://host[:porta], sem barra no final)", origin)
		}
	}
	v.atLeast("CORS_MAX_AGE", c.CORS.MaxAge, 0)

	if c.Mail.Enabled {
		v.required("MAIL_HOST", c.Mail.Host)
		v.port("MAIL_PORT", c.Mail.Port)
		if _, err := mail.ParseAddress(c.Mail.From); err != nil {
			v.addf("MAIL_FROM deve ser um e-mail válido")
		}
	}

	return v.problems
}

func isOrigin(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

type validator struct {
	problems []string
}

func (v *validator) addf(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

func (v *validator) required(key, value string) {
	if value == "" {
		v.addf("%s é obrigatória", key)
	}
}

func (v *validator) port(key string, value int) {
	if value < 1 || value > 65535 {
		v.addf("%s deve ser uma porta entre 1 e 65535 (recebido %d)", key, value)
	}
}

func (v *validator) atLeast(key string, value, minimum int) {
	if value < minimum {
		v.addf("%s deve ser no mínimo %d (recebido %d)", key, minimum, value)
	}
}

func (v *validator) notNegative(key string, value time.Duration) {
	if value < 0 {
		v.addf("%s não pode ser negativo", key)
	}
}

// envReader converte as variáveis de ambiente para os tipos do Config. Variável
// ausente usa o padrão. Variável definida com valor que não pode ser convertido
// (inclusive vazio, exceto em strings) vira um problema e impede a subida.
type envReader struct {
	problems []string
}

func (r *envReader) addf(format string, args ...any) {
	r.problems = append(r.problems, fmt.Sprintf(format, args...))
}

func (r *envReader) str(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func (r *envReader) boolean(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		r.addf("%s deve ser true ou false (recebido %q)", key, value)
		return fallback
	}
	return parsed
}

func (r *envReader) integer(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		r.addf("%s deve ser um número inteiro (recebido %q)", key, value)
		return fallback
	}
	return parsed
}

func (r *envReader) minutes(key string, fallback int) time.Duration {
	n := int64(r.integer(key, fallback))
	if limit := int64(math.MaxInt64 / time.Minute); n > limit || n < -limit {
		r.addf("%s está fora do intervalo permitido (recebido %d)", key, n)
		return time.Duration(fallback) * time.Minute
	}
	return time.Duration(n) * time.Minute
}

// list não aceita lista vazia: em CORS_ALLOWED_ORIGINS, cair no padrão "*"
// liberaria todas as origens por causa de um valor esquecido.
func (r *envReader) list(key, fallback string) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		value = fallback
	}

	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		r.addf("%s não pode ser vazia", key)
	}
	return items
}
