# Arquitetura do Backend

## Camadas

```
┌──────────────────────────────────────────────────────┐
│  nginx                                               │
│  /api/* → proxy_pass http://back:8080/ (sem /api)    │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  cmd/api/main.go                                     │
│  config.Load → database.Connect → router.Setup       │
│  http.Server com graceful shutdown (SIGTERM)         │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  providers (Boot) — toda requisição                  │
│  Logger: imprime método/url/path/ip (+ body opc.)    │
│  CORS:   headers + responde preflight OPTIONS        │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  middlewares — por rota                              │
│  RateLimit (token bucket por IP)                     │
│  Auth (Bearer JWT → claims no context)               │
│  RequireRole("admin")                                │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  domain/<x>/handler   — papel de controller          │
│  decodifica JSON, monta Actor, mapeia erro → HTTP    │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  domain/<x>/service   — orquestração                 │
│  define a ordem dos passos de cada operação          │
│  (chamado pelo handler e pelo seeder)                │
└──────────────────────┬───────────────────────────────┘
                       │ 1..N
┌──────────────────────▼───────────────────────────────┐
│  domain/<x>/usecase   — um passo por arquivo         │
│  função de responsabilidade única: autorizar,        │
│  validar, gravar, gerar token, enviar e-mail...      │
└──────────────────────┬───────────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────────┐
│  domain/<x>/repository — interface + GORM            │
└──────────────────────┬───────────────────────────────┘
                       │
                   MySQL 8.0
```

Na volta, o Logger (provider mais externo) imprime status e tempo de execução.

## Regras de dependência
- `handler` → `service` → `usecase` → `repository` → `model`
- `handler` nunca chama `usecase` diretamente; `usecase` nunca chama `service` nem outro `usecase` (quem combina os passos é o `service`)
- `errors` e `model` não importam nada do domínio
- `infra/*`, `response`, `config` não importam domínios
- `router` é o único lugar que conhece todas as peças (composição)

## Rotas atuais

| Método | Rota | Middlewares | Regra |
|---|---|---|---|
| GET | `/health` | — | pinga o MySQL (timeout 2s): 200 ou 503 com `details.database = "down"` |
| POST | `/auth/login` | rate_limit | |
| POST | `/auth/refresh` | rate_limit | Bearer com token válido ou expirado há ≤ `JWT_REFRESH_GRACE_MINUTES`; recarrega o usuário do banco |
| GET | `/auth/me` | rate_limit, auth | |
| GET | `/users?page=&perPage=` | rate_limit, auth, admin | |
| POST | `/users` | rate_limit, auth, admin | envia e-mail de boas-vindas |
| GET | `/users/{id}` | rate_limit, auth | admin ou o próprio usuário |
| PUT | `/users/{id}` | rate_limit, auth | admin ou o próprio; só admin muda `role` |
| DELETE | `/users/{id}` | rate_limit, auth, admin | soft delete; não remove a si mesmo |

## Migrations e seeders
`cmd/migration` usa `migration.Migrate` (GORM AutoMigrate dos `models`) e `migration.Seed`.
No docker-compose o serviço `migrate` roda `/app/migration -seed` e o `back` só sobe depois
que ele termina com sucesso.
