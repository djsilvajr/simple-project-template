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
    ├── config/config.go       Load(): env tipado e validado (valor inválido impede a subida)
    ├── domain/
    │   └── user/
    │       ├── errors/        erros de domínio (sentinelas + ValidationError)
    │       ├── handler/       "controller": lê request, chama service, escreve response
    │       ├── model/         entidade User + DTOs de entrada/saída
    │       ├── repository/    interface + implementação GORM
    │       ├── service/       orquestra cada operação chamando um ou mais use cases
    │       └── usecase/       funções de responsabilidade única (uma por arquivo)
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
      → handler             decodifica JSON, monta Actor, chama o service
        → service           define a ordem dos passos da operação
          → [use cases]     um passo cada: autorizar, validar, gravar, gerar token...
            → repository    GORM → MySQL
  ← providers (fim)         Logger imprime status e duração
```

## Service e use cases
O use case é uma função de responsabilidade única, em um arquivo próprio dentro de `domain/<x>/usecase/`, com as dependências recebidas por parâmetro. O service é quem chama os use cases, quantos a operação precisar:

```go
func (s *service) Delete(ctx context.Context, actor model.Actor, id uint) error {
	if err := usecase.AuthorizeAdmin(actor); err != nil {
		return err
	}
	if err := usecase.EnsureNotSelf(actor, id); err != nil {
		return err
	}
	return usecase.RemoveUser(ctx, s.repo, id)
}
```

- O handler chama apenas o service.
- Um use case não chama o service nem outro use case.
- Regras compartilhadas por vários use cases ficam em funções não exportadas do pacote (ex.: `validation_rules.go`).

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
Veja `back/.env.example` — todas comentadas. Obrigatórias: `DB_DATABASE`, `DB_USERNAME`, `JWT_SECRET` (mínimo de 32 bytes), `PASSWORD_SECRET` (e `MAIL_HOST` se `MAIL_ENABLED=true`).

`config.Load` é usado pela API e pelo executável de migrations. Se algum valor estiver incorreto, o processo encerra com erro listando todos os problemas de uma vez:
- Variável ausente usa o valor padrão, exceto as obrigatórias.
- Número ou booleano inválido ou vazio é erro, assim como lista vazia (ex.: `CORS_ALLOWED_ORIGINS=`).
- Faixas também são conferidas (portas, expiração do JWT, pool de conexões, origens de CORS). As regras de rate limit e de e-mail só valem com o recurso habilitado.

## Como adicionar um novo domínio
1. Crie `internal/domain/<nome>/` com `errors/`, `model/`, `repository/`, `service/`, `usecase/`, `handler/`
2. Adicione o model em `migration/migrate.go` (`models`)
3. Instancie repository → service → handler em `router.Setup`
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
