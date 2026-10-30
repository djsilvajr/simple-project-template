package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

type claimsKey struct{}

// Auth exige um header "Authorization: Bearer <jwt>" válido e coloca as
// claims no contexto da requisição (recuperáveis com ClaimsFromContext).
func Auth(tokens *token.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, found := BearerToken(r)
			if !found {
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				response.Error(w, http.StatusUnauthorized, "Token de acesso ausente.")
				return
			}

			claims, err := tokens.Parse(raw)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer realm="api", error="invalid_token"`)
				response.Error(w, http.StatusUnauthorized, "Token inválido ou expirado.")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole permite apenas usuários com um dos papéis informados.
// Deve ser usado depois de Auth.
func RequireRole(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Não autenticado.")
				return
			}
			if !slices.Contains(roles, claims.Role) {
				response.Error(w, http.StatusForbidden, "Sem permissão para acessar este recurso.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BearerToken extrai o token do header "Authorization: Bearer <token>".
func BearerToken(r *http.Request) (string, bool) {
	raw, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	raw = strings.TrimSpace(raw)
	return raw, found && raw != ""
}

func ClaimsFromContext(ctx context.Context) (*token.Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(*token.Claims)
	return claims, ok
}
