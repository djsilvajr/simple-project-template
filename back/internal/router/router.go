package router

import (
	"context"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	userhandler "github.com/djsilvajr/simple-project-template/back/internal/domain/user/handler"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	userrepository "github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
	userservice "github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/mailer"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
	"github.com/djsilvajr/simple-project-template/back/internal/middleware"
	"github.com/djsilvajr/simple-project-template/back/internal/providers"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

type Dependencies struct {
	Config *config.Config
	DB     *gorm.DB
	Mailer mailer.Mailer
}

// Setup monta a aplicação HTTP. Fluxo de cada requisição:
//
//	providers (Boot) -> middlewares da rota -> handler -> service -> [use cases] -> providers (fim)
func Setup(deps Dependencies) (http.Handler, error) {
	cfg := deps.Config

	// 1. Providers: Register roda uma vez aqui; Boot envolve todas as rotas.
	appProviders := providers.Default()
	if err := providers.RegisterAll(cfg, appProviders); err != nil {
		return nil, err
	}

	// 2. Dependências dos domínios.
	tokens := token.NewManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiration, cfg.App.Name)

	users := userhandler.New(userservice.New(userservice.Deps{
		Repo:           userrepository.New(deps.DB),
		PasswordSecret: cfg.Auth.PasswordSecret,
		Tokens:         tokens,
		RefreshGrace:   cfg.Auth.JWTRefreshGrace,
		Mailer:         deps.Mailer,
	}))

	// 3. Middlewares.
	rateLimit := middleware.RateLimit(cfg)
	auth := middleware.Auth(tokens)
	admin := middleware.RequireRole(string(model.RoleAdmin))

	// 4. Rotas (sem o prefixo /api — o nginx remove o prefixo antes de repassar).
	mux := http.NewServeMux()
	route := func(pattern string, h http.HandlerFunc, mws ...middleware.Middleware) {
		mux.Handle(pattern, middleware.Chain(h, mws...))
	}

	route("GET /health", health(deps.DB))

	route("POST /auth/login", users.Login, rateLimit)
	route("POST /auth/refresh", users.Refresh, rateLimit) // valida o token por conta própria (aceita expirado dentro da janela)
	route("GET /auth/me", users.Me, rateLimit, auth)

	route("GET /users", users.List, rateLimit, auth, admin)
	route("POST /users", users.Create, rateLimit, auth, admin)
	route("GET /users/{id}", users.Get, rateLimit, auth)
	route("PUT /users/{id}", users.Update, rateLimit, auth)
	route("DELETE /users/{id}", users.Delete, rateLimit, auth, admin)

	route("/", notFound)

	return providers.Wrap(mux, appProviders), nil
}

const healthDBTimeout = 2 * time.Second

// health verifica se a API e o banco estão respondendo.
// 200 quando tudo está ok; 503 quando o banco não responde.
func health(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := map[string]string{"api": "ok", "database": "ok"}

		if err := pingDB(r.Context(), db); err != nil {
			log.Printf("health: banco indisponível: %v", err)
			checks["database"] = "down"
			response.ErrorWithDetails(w, http.StatusServiceUnavailable, "Serviço indisponível.", checks)
			return
		}

		response.Success(w, http.StatusOK, checks)
	}
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, healthDBTimeout)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotFound, "Rota não encontrada.")
}
