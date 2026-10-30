# Contexto do Backend

## Stack
- Go 1.22 — `net/http` da biblioteca padrão (roteamento com padrões `"METHOD /path/{id}"`)
- MySQL 8.0 via GORM (`gorm.io/gorm` + `gorm.io/driver/mysql`)
- JWT HS256 (`golang-jwt/jwt/v5`) para autenticação
- bcrypt + pepper HMAC-SHA256 (`PASSWORD_SECRET`) para senhas
- `joho/godotenv` para carregar o `.env` (variáveis do ambiente têm prioridade)
- Módulo: `github.com/djsilvajr/simple-project-template/back`

## Estrutura
```
back/
├── cmd/
│   ├── api/main.go            executável da API (HTTP + graceful shutdown)
│   └── migration/main.go      executável de migrations/seeders
├── docker/Dockerfile          build multi-stage → imagem com /app/api e /app/migration
├── migration/
│   ├── migrate.go             AutoMigrate dos models (lista `models`)
│   └── seeder.go              seeders idempotentes (lista `seeders`)
└── internal/
    ├── config/config.go       Load() + getEnvStr / getEnvBool / getEnvInt
    ├── domain/
    │   └── user/
    │       ├── errors/        erros de domínio (sentinelas + ValidationError)
    │       ├── handler/       "controller": lê request, chama use case, escreve response
    │       ├── model/         entidade User + DTOs de entrada/saída
    │       ├── repository/    interface + implementação GORM
    │       ├── service/       regras de domínio reutilizáveis (validação, unicidade, hash)
    │       └── usecase/       fluxos da aplicação + autorização (um arquivo por caso de uso)
    ├── infra/
    │   ├── database/          conexão GORM + pool + retry
    │   ├── mailer/            SMTP (ou log quando MAIL_ENABLED=false)
    │   └── token/             geração/validação de JWT
    ├── middleware/            executam por rota, antes do handler (auth, rate_limit)
    ├── providers/             executam em TODA requisição, no início e no fim (logger, cors)
    ├── response/              envelope JSON padrão (Success / Error)
    └── router/router.go       monta providers → middlewares → rotas
```

## Fluxo de uma requisição
```
nginx /api/users → proxy_pass http://back:8080/users   (prefixo /api removido)
  → providers (Boot)        Logger (início) → CORS
    → middlewares da rota   RateLimit → Auth → RequireRole
      → handler             decodifica JSON, monta Actor, chama use case
        → usecase           autorização (admin / próprio usuário) + orquestração
          → service         validação, unicidade, hash de senha
            → repository    GORM → MySQL
  ← providers (fim)         Logger imprime status e duração
```

## Providers vs. middlewares
| | Providers (`internal/providers`) | Middlewares (`internal/middleware`) |
|---|---|---|
| Quando | Toda requisição, inclusive 404 e preflight | Apenas nas rotas onde são declarados |
| Ordem | Antes de qualquer middleware/rota e depois de tudo | Entre o provider e o handler |
| Interface | `Register(cfg)` (1x no startup) + `Boot(next)` (por requisição) | `func(http.Handler) http.Handler` |
| Registro | `providers.Default()` | `route(pattern, handler, mws...)` em `router.go` |

## Padrão de resposta (`internal/response`)
```go
response.Success(w, http.StatusOK, data)                     // { status: true, statusCode, data }
response.Error(w, http.StatusNotFound, "mensagem")           // { status: false, statusCode, error }
response.ErrorWithDetails(w, 422, "Dados inválidos.", campos) // + details
response.NoContent(w)                                         // 204
```
Listagens usam `response.Paginated{Items, Total, Page, PerPage}` dentro de `data`.

## Erros → HTTP (feito em `handler.writeError`)
| Erro de domínio | HTTP |
|---|---|
| `*ValidationError` | 422 (com `details` por campo) |
| `ErrInvalidCredentials`, `ErrSessionExpired` | 401 |
| `ErrForbidden`, `ErrCannotDeleteSelf` | 403 |
| `ErrNotFound` | 404 |
| `ErrLoginTaken`, `ErrEmailTaken` | 409 |
| qualquer outro | 500 (logado, mensagem genérica) |

## Variáveis de ambiente
Veja `back/.env.example` — todas comentadas. Obrigatórias: `DB_DATABASE`, `DB_USERNAME`, `JWT_SECRET`, `PASSWORD_SECRET` (e `MAIL_HOST` se `MAIL_ENABLED=true`).

## Como adicionar um novo domínio
1. Crie `internal/domain/<nome>/` com `errors/`, `model/`, `repository/`, `service/`, `usecase/`, `handler/`
2. Adicione o model em `migration/migrate.go` (`models`)
3. Instancie repository → service → use cases → handler em `router.Setup`
4. Registre as rotas em `router.Setup` com os middlewares necessários

Referência: `internal/domain/user`.

## Comandos
```bash
go run ./cmd/api                      # sobe a API (lê back/.env)
go run ./cmd/migration                # aplica migrations
go run ./cmd/migration -seed          # migrations + seeders
go run ./cmd/migration -seed-only     # apenas seeders
go vet ./... && go build ./...
```
