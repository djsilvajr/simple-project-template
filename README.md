# Simple Project Template

Template full-stack com arquitetura limpa, pronto para Docker. Backend em Go (DDD, `net/http` + GORM), frontend em HTML/CSS/JS/PHP, separados por nginx, com MySQL e Redis.

---

## Stack

| Camada | Tecnologia |
|---|---|
| Proxy reverso | nginx 1.25 |
| API | Go 1.22 · net/http · GORM (MySQL) · JWT |
| Frontend | HTML5 · CSS3 · JS ES2022 · PHP 8.2-FPM |
| Banco de dados | MySQL 8.0 |
| Cache | Redis 7 |
| Orquestração | Docker Compose |

---

## Estrutura do projeto

```
.
├── back/                       API Go
│   ├── cmd/
│   │   ├── api/                executável da API
│   │   └── migration/          executável de migrations/seeders
│   ├── docker/Dockerfile       build multi-stage (api + migration)
│   ├── migration/              migrate.go e seeder.go
│   ├── internal/
│   │   ├── config/             carregamento do .env
│   │   ├── domain/user/        errors, handler, model, repository, service, usecase
│   │   ├── infra/              database, mailer, token (JWT)
│   │   ├── middleware/         auth, rate_limit
│   │   ├── providers/          logger, cors (Register + Boot)
│   │   ├── response/           envelope JSON padrão
│   │   └── router/             providers → middlewares → rotas
│   ├── go.mod
│   └── .env.example
├── front/                      Frontend
│   ├── Auth/                   páginas de autenticação
│   ├── test/                   páginas de exemplo (pública e protegida)
│   ├── css/  js/               estilos e scripts por feature
│   ├── php/class/              sessão e proxy para API
│   └── .env.example
├── nginx/default.conf          /api/* → back:8080 · /* → front
├── docker/front/               Dockerfile PHP-FPM do front
├── docker-compose.yml
├── .env.example
└── llm/                        documentação para uso com IA
```

---

## Primeiros passos

### 1. Copiar e preencher os arquivos de ambiente

```bash
cp .env.example .env               # docker-compose (portas, credenciais do MySQL)
cp back/.env.example back/.env     # API Go
cp front/.env.example front/.env   # PHP do front
```

As credenciais de banco em `back/.env` (`DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`) precisam bater com as do `.env` da raiz. Troque `JWT_SECRET` e `PASSWORD_SECRET` por valores longos e aleatórios.

### 2. Subir os containers

```bash
docker compose up -d --build
```

O serviço `migrate` roda as migrations e os seeders (`/app/migration -seed`) e encerra; a API só sobe depois que ele termina com sucesso.

### 3. Acessar

| URL | Descrição |
|---|---|
| `http://localhost/` | Página inicial (pública) |
| `http://localhost/Auth/login.html` | Tela de login |
| `http://localhost/test/auth-needed.html` | Exemplo de página protegida |
| `http://localhost/api/health` | Health check da API |

---

## Usuário admin inicial

Criado pelo seeder a partir de `back/.env`:

| Variável | Padrão do `.env.example` |
|---|---|
| `SEED_ADMIN_LOGIN` | `admin` |
| `SEED_ADMIN_PASSWORD` | `senha12345` |
| `SEED_ADMIN_EMAIL` | `admin@example.com` |

O seeder é idempotente: se o login já existe, nada é alterado. Com `SEED_ADMIN_PASSWORD` vazio, o admin não é criado.

> ⚠️ Senhas são guardadas como `bcrypt(HMAC-SHA256(PASSWORD_SECRET, senha))`. **Trocar o `PASSWORD_SECRET` invalida todas as senhas existentes.**

---

## Rotas da API

Base: `http://localhost/api` · Autenticação: `Authorization: Bearer <token>`

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| `GET` | `/health` | — | Health check da API e do banco (200 ok · 503 banco fora) |
| `POST` | `/auth/login` | — | Login com `login` e `password` |
| `POST` | `/auth/refresh` | token (válido ou expirado há ≤ 30 min) | Renova o token |
| `GET` | `/auth/me` | token | Usuário autenticado |
| `GET` | `/users?page=1&perPage=20` | admin | Listar usuários |
| `POST` | `/users` | admin | Criar usuário |
| `GET` | `/users/{id}` | admin ou o próprio | Buscar usuário |
| `PUT` | `/users/{id}` | admin ou o próprio | Atualizar (só admin altera `role`) |
| `DELETE` | `/users/{id}` | admin | Remover (soft delete) |

### POST `/api/auth/login`

```json
// Request
{ "login": "admin", "password": "senha12345" }

// Response 200
{
  "status": true,
  "statusCode": 200,
  "data": {
    "token": "eyJhbGciOi...",
    "expiresAt": 1798650000,
    "user": { "id": 1, "login": "admin", "name": "Administrador", "email": "admin@example.com", "role": "admin", "createdAt": "...", "updatedAt": "..." }
  }
}

// Response 401
{ "status": false, "statusCode": 401, "error": "login ou senha inválidos" }
```

### POST `/api/users`

```json
// Request
{ "login": "joao", "name": "João", "email": "joao@email.com", "password": "minhasenha", "role": "user" }

// Response 201
{ "status": true, "statusCode": 201, "data": { "id": 2, "login": "joao", ... } }

// Response 422 — validação
{ "status": false, "statusCode": 422, "error": "Dados inválidos.", "details": { "email": "formato inválido" } }

// Response 409 — conflito
{ "status": false, "statusCode": 409, "error": "login já cadastrado" }
```

---

## Controle de sessão no frontend

O login é feito com **login e senha**. O PHP do front (`php/class/Auth/auth.php`) chama a API e guarda `user` e o JWT (`token`) em `$_SESSION`.

Quando o token expira, a próxima requisição do usuário ao PHP do front tenta renová-lo em `POST /api/auth/refresh`. Isso só funciona se o token expirou há no máximo 30 minutos (`JWT_REFRESH_GRACE_MINUTES` / `SESSION_REFRESH_GRACE`). Se a renovação falhar, a sessão é encerrada e o usuário volta ao login.

**Página protegida** — inclua `session-check.js` como primeiro script no `<head>`:

```html
<head>
    <script src="/js/Auth/session-check.js"></script>
</head>
```

Para chamar rotas protegidas da API a partir do PHP do front:

```php
$curl = new CurlHelper($_SESSION['token']);
$resp = $curl->get('/auth/me');
```

---

## Adicionando um novo domínio no backend

Veja `back/context.md`. Resumo:

```
back/internal/domain/<nome>/
├── errors/      erros de domínio
├── model/       entidade + DTOs
├── repository/  interface + GORM
├── service/     regras de domínio
├── usecase/     um arquivo por caso de uso
└── handler/     leitura do request / escrita do response
```

Depois: adicione o model em `back/migration/migrate.go` e registre as rotas em `back/internal/router/router.go`.

---

## Comandos úteis

```bash
# Subir tudo
docker compose up -d --build

# Logs
docker compose logs -f back
docker compose logs migrate

# Rodar migrations/seeders manualmente
docker compose run --rm migrate                         # migrations + seed
docker compose run --rm migrate /app/migration          # só migrations

# Recriar banco do zero (apaga todos os dados)
docker compose down -v && docker compose up -d --build

# Desenvolvimento local (fora do Docker), dentro de back/
go run ./cmd/migration -seed
go run ./cmd/api
```

---

## Documentação interna

| Arquivo | Conteúdo |
|---|---|
| `back/context.md` | Stack, estrutura, padrões e comandos do backend |
| `back/archtecture.md` | Camadas, regras de dependência e rotas |
| `front/context.md` | Separação de responsabilidades e convenções do frontend |
| `front/archtecture.md` | Fluxo de autenticação e organização de arquivos |
| `llm/context/context.md` | Visão geral do projeto para uso com IA |
| `llm/context/archtecture.md` | Arquitetura completa de infraestrutura e camadas |
| `llm/task/task.example.md` | Templates e exemplos de tasks para desenvolvimento com IA |
