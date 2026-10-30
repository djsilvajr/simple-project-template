package token

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	m := NewManager("secret", time.Hour, "api")

	raw, _, err := m.Generate(7, "admin")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := m.Parse(raw)
	if err != nil {
		t.Fatalf("token válido rejeitado: %v", err)
	}
	if claims.UserID != 7 || claims.Role != "admin" {
		t.Fatalf("claims inesperadas: %+v", claims)
	}

	if _, err := NewManager("outro", time.Hour, "api").Parse(raw); err == nil {
		t.Fatal("token com assinatura de outra chave foi aceito")
	}
	if _, err := NewManager("secret", time.Hour, "outro").Parse(raw); err == nil {
		t.Fatal("token de outro emissor foi aceito")
	}
}

func TestParseForRefresh(t *testing.T) {
	const grace = 30 * time.Minute

	tests := []struct {
		name      string
		expiresIn time.Duration // negativo = já expirado
		parseOK   bool
		refreshOK bool
	}{
		{"válido", time.Hour, true, true},
		{"expirado há 10min", -10 * time.Minute, false, true},
		{"expirado há 29min", -29 * time.Minute, false, true},
		{"expirado há 31min", -31 * time.Minute, false, false},
		{"expirado há 2h", -2 * time.Hour, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _, err := NewManager("secret", tt.expiresIn, "api").Generate(1, "user")
			if err != nil {
				t.Fatal(err)
			}

			m := NewManager("secret", time.Hour, "api")

			if _, err := m.Parse(raw); (err == nil) != tt.parseOK {
				t.Errorf("Parse: ok=%v, esperado %v", err == nil, tt.parseOK)
			}
			if _, err := m.ParseForRefresh(raw, grace); (err == nil) != tt.refreshOK {
				t.Errorf("ParseForRefresh: ok=%v, esperado %v", err == nil, tt.refreshOK)
			}
		})
	}
}
