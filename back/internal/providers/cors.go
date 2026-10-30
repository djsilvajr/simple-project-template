package providers

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
)

// CORS adiciona os headers de CORS em toda resposta e encerra as
// requisições de preflight (OPTIONS) antes de chegarem aos middlewares/rotas.
type CORS struct {
	allowedOrigins   []string
	allowAll         bool
	allowedMethods   string
	allowedHeaders   string
	allowCredentials bool
	maxAge           string
}

func (c *CORS) Register(cfg *config.Config) error {
	c.allowedOrigins = cfg.CORS.AllowedOrigins
	c.allowAll = slices.Contains(cfg.CORS.AllowedOrigins, "*")
	c.allowedMethods = strings.Join(cfg.CORS.AllowedMethods, ", ")
	c.allowedHeaders = strings.Join(cfg.CORS.AllowedHeaders, ", ")
	c.allowCredentials = cfg.CORS.AllowCredentials
	c.maxAge = strconv.Itoa(cfg.CORS.MaxAge)
	return nil
}

func (c *CORS) Boot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" && c.isAllowed(origin) {
			h := w.Header()
			h.Add("Vary", "Origin")

			// Com credenciais o navegador não aceita "*", então ecoa a origem.
			if c.allowAll && !c.allowCredentials {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
			}
			if c.allowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}

			if isPreflight(r) {
				h.Set("Access-Control-Allow-Methods", c.allowedMethods)
				h.Set("Access-Control-Allow-Headers", c.allowedHeaders)
				h.Set("Access-Control-Max-Age", c.maxAge)
			}
		}

		if isPreflight(r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *CORS) isAllowed(origin string) bool {
	return c.allowAll || slices.Contains(c.allowedOrigins, origin)
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
