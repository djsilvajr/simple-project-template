# Simple Project Template

Template PHP full-stack com arquitetura limpa, pronto para Docker. Backend em PHP puro com DDD, frontend em HTML/CSS/JS/PHP, separados por nginx, com MySQL e Redis.

---

## Stack

| Camada | Tecnologia |
|---|---|
| Proxy reverso | nginx 1.25 |
| API | PHP 8.2-FPM · PDO MySQL · phpdotenv v5 |
| Frontend | HTML5 · CSS3 · JS ES2022 · PHP 8.2-FPM |
| Banco de dados | MySQL 8.0 |
| Cache | Redis 7 |
| Orquestração | Docker Compose |

---

## Estrutura do projeto

```
.
├── back/                   API PHP (DDD)
│   ├── src/
│   │   ├── Core/           roteador, PDO, middleware, response
│   │   ├── Http/           registro de rotas
│   │   ├── Routes/         main.php — ponto central de rotas
│   │   └── Domain/
│   │       └── Usuario/    domínio de usuários (Controller, UseCase, Repository, DTO)
│   ├── composer.json
│   ├── index.php
│   └── .env.exemple
├── front/                  Frontend
│   ├── Auth/               páginas de autenticação
│   ├── test/               páginas de exemplo (pública e protegida)
│   ├── css/                estilos por feature
│   ├── js/                 scripts por feature
│   └── php/class/          sessão e proxy para API
├── nginx/
│   └── default.conf        roteamento /api/* → back · /* → front
├── docker/
│   ├── back/               Dockerfile PHP-FPM + Composer
│   └── front/              Dockerfile PHP-FPM
├── db/
│   └── create_usuarios.sql tabela + usuário de teste
├── docker-compose.yml
├── .env.example
└── llm/                    documentação para uso com IA
```

---

## Primeiros passos

### 1. Copiar e preencher os arquivos de ambiente

```bash
cp .env.example .env
cp back/.env.exemple back/.env
```

**`.env`** (raiz) — usado pelo docker-compose:

```env
PROJECT_NAME=myapp
NGINX_PORT=80
DB_ROOT_PASSWORD=rootpassword
DB_DATABASE=mydb
DB_USERNAME=myuser
DB_PASSWORD=mypassword
DB_PORT_EXPOSE=3306
REDIS_PORT_EXPOSE=6379
```

**`back/.env`** — usado pelo PHP da API:

```env
PROJECT_NAME=api
DB_HOST=mysql
DB_PORT=3306
DB_DATABASE=mydb
DB_USERNAME=myuser
DB_PASSWORD=mypassword
API_USER=admin
API_PASSWORD=secret
PASSWORD_SECRET=t3mpl@t3-s3cr3t-k3y-2024
REDIS_HOST=redis
REDIS_PORT=6379
```

> **Importante:** `PROJECT_NAME=api` no `back/.env` deve permanecer `api` para o roteador remover o prefixo `/api` das URLs corretamente.

### 2. Subir os containers

```bash
docker compose up -d --build
```

O MySQL executa automaticamente os scripts de `db/` na primeira inicialização, criando a tabela `usuarios` e inserindo o usuário de teste.

### 3. Acessar

| URL | Descrição |
|---|---|
| `http://localhost/` | Página inicial (pública) |
| `http://localhost/Auth/login.html` | Tela de login |
| `http://localhost/test/auth-needed.html` | Exemplo de página protegida |
| `http://localhost/test/non-auth-needed.html` | Exemplo de página pública |
| `http://localhost/api/` | Endpoint da API |

---

## Usuário de teste

> **Criado automaticamente ao subir o banco pela primeira vez.**

| Campo | Valor |
|---|---|
| usuario | `admin` |
| email | `admin@teste.com` |
| senha | `senha123` |

O hash armazenado no banco foi gerado com `password_hash('senha123' . PASSWORD_SECRET, PASSWORD_BCRYPT)`, onde `PASSWORD_SECRET=t3mpl@t3-s3cr3t-k3y-2024`.

> ⚠️ **Se você trocar o `PASSWORD_SECRET` no `back/.env`**, o hash do usuário de teste deixa de ser válido. Gere um novo hash e atualize `db/create_usuarios.sql`:
>
> ```bash
> php -r "echo password_hash('senha123' . 'SEU_NOVO_SECRET', PASSWORD_BCRYPT);"
> ```
>
> Depois recrie o volume do banco:
>
> ```bash
> docker compose down -v
> docker compose up -d --build
> ```

---

## Rotas da API

Base: `http://localhost/api`

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| `POST` | `/auth/login` | Autenticar usuário | — |
| `POST` | `/usuarios` | Criar novo usuário | — |

### POST `/api/auth/login`

```json
// Request
{ "email": "admin@teste.com", "senha": "senha123" }

// Response 200
{
  "status": true,
  "statusCode": 200,
  "data": { "id": 1, "usuario": "admin", "email": "admin@teste.com", "criacao_datahora": "...", "alteracao_datahora": "..." }
}

// Response 401
{ "status": false, "statusCode": 401, "error": "Credenciais inválidas." }
```

### POST `/api/usuarios`

```json
// Request
{ "usuario": "joao", "email": "joao@email.com", "senha": "minhasenha" }

// Response 201
{
  "status": true,
  "statusCode": 201,
  "data": { "id": 2, "usuario": "joao", "email": "joao@email.com", "criacao_datahora": "..." }
}

// Response 422 — validação
{ "status": false, "statusCode": 422, "error": "Email inválido." }

// Response 409 — conflito
{ "status": false, "statusCode": 409, "error": "Email já cadastrado." }
```

---

## Controle de sessão no frontend

A sessão é gerenciada pelo PHP do container front via `$_SESSION`.

**Página pública** — nenhuma configuração necessária.

**Página protegida** — inclua `session-check.js` como primeiro script no `<head>`:

```html
<head>
    <script src="/js/Auth/session-check.js"></script>
</head>
```

O script oculta a página, verifica a sessão e redireciona para `/Auth/login.html` se não autenticado. Após confirmação, expõe `window.__session` com os dados do usuário.

---

## Adicionando um novo domínio no backend

Crie a estrutura em `back/src/Domain/NomeDominio/`:

```
NomeDominio/
├── Controller/NomeDominioController.php
├── UseCase/NomeUseCase.php
├── Repository/
│   ├── NomeDominioRepositoryInterface.php
│   └── NomeDominioRepository.php
└── DTO/NomeDominioDTO.php
```

Registre as rotas em `back/src/Routes/main.php`:

```php
Route::get('/recurso',      'Domain\\NomeDominio\\Controller\\NomeDominioController@listar', []);
Route::post('/recurso',     'Domain\\NomeDominio\\Controller\\NomeDominioController@criar',  []);
Route::put('/recurso/{id}', 'Domain\\NomeDominio\\Controller\\NomeDominioController@editar', ['auth']);
```

Consulte `Domain/Teste` como referência de implementação completa e `llm/task/task.example.md` para usar IA no desenvolvimento.

---

## Comandos úteis

```bash
# Subir tudo
docker compose up -d --build

# Ver logs de um serviço
docker compose logs -f back
docker compose logs -f nginx

# Recriar banco do zero (apaga todos os dados)
docker compose down -v && docker compose up -d --build

# Acessar o container do back
docker compose exec back sh

# Rodar composer manualmente
docker compose exec back composer install
```

---

## Documentação interna

| Arquivo | Conteúdo |
|---|---|
| `back/context.md` | Stack, variáveis de ambiente e padrões do backend |
| `back/archtecture.md` | Diagrama de camadas e ciclo de vida de uma requisição |
| `front/context.md` | Separação de responsabilidades e convenções do frontend |
| `front/archtecture.md` | Fluxo de autenticação e organização de arquivos |
| `llm/context/context.md` | Visão geral do projeto para uso com IA |
| `llm/context/archtecture.md` | Arquitetura completa de infraestrutura e camadas |
| `llm/task/task.example.md` | Templates e exemplos de tasks para desenvolvimento com IA |
