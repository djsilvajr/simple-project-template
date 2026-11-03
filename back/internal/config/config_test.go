package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

// setEnv troca o ambiente inteiro pelo conjunto mínimo válido + overrides.
func setEnv(t *testing.T, overrides map[string]string) {
	t.Helper()

	saved := os.Environ()
	t.Cleanup(func() {
		os.Clearenv()
		for _, kv := range saved {
			key, value, _ := strings.Cut(kv, "=")
			os.Setenv(key, value)
		}
	})
	os.Clearenv()

	env := map[string]string{
		"DB_DATABASE":     "db",
		"DB_USERNAME":     "user",
		"JWT_SECRET":      strings.Repeat("x", minJWTSecretBytes),
		"PASSWORD_SECRET": "pepper",
	}
	for key, value := range overrides {
		env[key] = value
	}
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func TestLoadValidEnv(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_PORT":                  " 9000 ",
		"TRUST_PROXY":               "true",
		"JWT_EXPIRATION_MINUTES":    "15",
		"JWT_REFRESH_GRACE_MINUTES": "0",
		"CORS_ALLOWED_ORIGINS":      "http://localhost:8088, https://app.example.com",
		"RATE_LIMIT_ENABLED":        "false",
		"RATE_LIMIT_BURST":          "0",
		"MAIL_HOST":                 "",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.App.Port != 9000 || !cfg.App.TrustProxy {
		t.Errorf("App = %+v", cfg.App)
	}
	if cfg.Auth.JWTExpiration != 15*time.Minute || cfg.Auth.JWTRefreshGrace != 0 {
		t.Errorf("Auth: expiração=%s janela=%s", cfg.Auth.JWTExpiration, cfg.Auth.JWTRefreshGrace)
	}
	if got := cfg.CORS.AllowedOrigins; len(got) != 2 || got[1] != "https://app.example.com" {
		t.Errorf("AllowedOrigins = %v", got)
	}
}

func TestLoadUsesDefaultsWhenAbsent(t *testing.T) {
	setEnv(t, nil)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.App.Port != 8080 || cfg.Auth.JWTExpiration != time.Hour || cfg.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("padrões não aplicados: %+v", cfg)
	}
}

func TestLoadRejectsInvalidEnv(t *testing.T) {
	cases := []struct{ key, value string }{
		{"APP_NAME", ""},
		{"APP_PORT", "abc"},
		{"APP_PORT", ""},
		{"APP_PORT", "0"},
		{"APP_PORT", "70000"},
		{"TRUST_PROXY", "ture"},
		{"TRUST_PROXY", ""},
		{"DB_HOST", ""},
		{"DB_PORT", "-1"},
		{"DB_DATABASE", ""},
		{"DB_USERNAME", "  "},
		{"DB_MAX_OPEN_CONNS", "0"},
		{"DB_MAX_IDLE_CONNS", "-1"},
		{"DB_MAX_IDLE_CONNS", "26"},
		{"DB_CONN_MAX_LIFETIME_MINUTES", "-1"},
		{"JWT_SECRET", "curta"},
		{"PASSWORD_SECRET", ""},
		{"JWT_EXPIRATION_MINUTES", "0"},
		{"JWT_EXPIRATION_MINUTES", "1h"},
		{"JWT_EXPIRATION_MINUTES", "999999999999"},
		{"JWT_REFRESH_GRACE_MINUTES", "-5"},
		{"RATE_LIMIT_BURST", "0"},
		{"RATE_LIMIT_REQUESTS_PER_SECOND", "0"},
		{"CORS_ALLOWED_ORIGINS", ""},
		{"CORS_ALLOWED_ORIGINS", "http://app.example.com/"},
		{"CORS_ALLOWED_ORIGINS", "app.example.com"},
		{"CORS_ALLOWED_METHODS", " , "},
		{"CORS_MAX_AGE", "-1"},
	}

	for _, tc := range cases {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			setEnv(t, map[string]string{tc.key: tc.value})

			if _, err := Load(); err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Errorf("esperava erro citando %s, veio: %v", tc.key, err)
			}
		})
	}
}

func TestLoadValidatesMailOnlyWhenEnabled(t *testing.T) {
	bad := map[string]string{"MAIL_PORT": "0", "MAIL_FROM": "invalido"}

	setEnv(t, bad)
	if _, err := Load(); err != nil {
		t.Errorf("e-mail desabilitado não deveria validar: %v", err)
	}

	bad["MAIL_ENABLED"] = "true"
	setEnv(t, bad)
	_, err := Load()
	if err == nil {
		t.Fatal("esperava erro")
	}
	for _, key := range []string{"MAIL_HOST", "MAIL_PORT", "MAIL_FROM"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("erro não cita %s: %v", key, err)
		}
	}
}

func TestLoadReportsAllProblemsWithoutLeakingSecrets(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_PORT":   "abc",
		"JWT_SECRET": "segredo-curto",
		"DB_PORT":    "0",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("esperava erro")
	}
	for _, key := range []string{"APP_PORT", "JWT_SECRET", "DB_PORT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("erro não cita %s: %v", key, err)
		}
	}
	if strings.Contains(err.Error(), "segredo-curto") {
		t.Errorf("erro expõe o segredo: %v", err)
	}
}
