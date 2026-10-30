package middleware

import "net/http"

// Middleware executa antes do handler de uma rota específica e pode
// interromper o fluxo (ex.: 401, 429) sem chamar o próximo.
type Middleware func(http.Handler) http.Handler

// Chain aplica os middlewares na ordem em que foram informados:
// Chain(h, a, b) executa a -> b -> h.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
