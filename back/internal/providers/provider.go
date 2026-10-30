package providers

import (
	"net/http"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
)

// Provider segue a ideia dos service providers do Laravel:
//
//   - Register: executa UMA vez, na subida da aplicação (configuração,
//     preparação de recursos).
//   - Boot: executa em TODA requisição, envolvendo o fluxo inteiro. O código
//     antes de next.ServeHTTP roda no início (antes de qualquer middleware e
//     rota) e o código depois roda no fim do fluxo.
type Provider interface {
	Register(cfg *config.Config) error
	Boot(next http.Handler) http.Handler
}

// Default é a lista de providers da aplicação, na ordem de execução.
// O primeiro da lista é o mais externo (inicia primeiro e finaliza por último).
func Default() []Provider {
	return []Provider{
		&Logger{},
		&CORS{},
	}
}

// RegisterAll chama Register de cada provider.
func RegisterAll(cfg *config.Config, providers []Provider) error {
	for _, p := range providers {
		if err := p.Register(cfg); err != nil {
			return err
		}
	}
	return nil
}

// Wrap envolve o handler com o Boot de cada provider, mantendo a ordem da lista.
func Wrap(h http.Handler, providers []Provider) http.Handler {
	for i := len(providers) - 1; i >= 0; i-- {
		h = providers[i].Boot(h)
	}
	return h
}
