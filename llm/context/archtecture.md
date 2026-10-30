# Arquitetura Geral do Projeto

## Diagrama de infraestrutura

```
                        ┌──────────────────────────────────┐
                        │  Browser                         │
                        └──────────────┬───────────────────┘
                                       │ HTTP :80
                        ┌──────────────▼───────────────────┐
                        │  nginx (proxy reverso)           │
                        │                                  │
                        │  /api/*  → back:8080             │
                        │  /*.php  → front:9000            │
                        │  /*      → static /var/www/front │
                        └────────┬──────────────┬──────────┘
                                 │              │
              ┌──────────────────▼──┐    ┌──────▼──────────────────┐
              │  back (Go)          │    │  front (PHP-FPM)        │
              │  :8080              │    │  :9000                  │
              │                    │    │                          │
              │  DDD API           │    │  Sessão + cURL          │
              └────────┬───────────┘    └──────────────────────────┘
                       │
          ┌────────────┼─────────────┐
          │            │             │
   ┌──────▼──────┐  ┌──▼──────┐  ┌──▼──────┐
   │  MySQL 8.0  │  │ Redis 7 │  │  (ext.) │
   └─────────────┘  └─────────┘  └─────────┘
```

## Camadas do sistema

### 1. Infraestrutura (Docker)
- **nginx**: único ponto de entrada externo. Roteamento baseado em prefixo de URL.
- **back**: binário Go (`/app/api`) servindo HTTP na porta 8080, acessível só pela rede interna.
- **migrate**: mesma imagem do back; roda `/app/migration -seed` e encerra antes do back subir.
- **front**: PHP-FPM para arquivos `.php` do frontend. Arquivos estáticos são servidos pelo nginx diretamente.
- **mysql**: banco relacional. Schema criado pelo serviço `migrate`.
- **redis**: disponível para cache, filas ou sessões distribuídas.

### 2. Backend (back/)
Go com arquitetura DDD (detalhes em `back/context.md` e `back/archtecture.md`):
```
cmd/api, cmd/migration →  executáveis
internal/config        →  Load() do .env
internal/providers     →  executam em toda requisição (logger, cors)
internal/middleware    →  executam por rota (rate_limit, auth)
internal/router        →  providers → middlewares → handler
internal/response      →  envelope JSON padrão
internal/infra         →  database (GORM), mailer (SMTP), token (JWT)
internal/domain/<x>    →  lógica de negócio por domínio
  handler    →  papel de controller: lê request, escreve response
  usecase    →  fluxos da aplicação + autorização
  service    →  regras de domínio (validação, unicidade, hash)
  repository →  acesso a dados (GORM)
  model      →  entidade + DTOs
  errors     →  erros de domínio
migration/             →  migrate.go e seeder.go
```

### 3. Frontend (front/)
Separação de responsabilidades por tipo de arquivo:
```
HTML  →  estrutura e conteúdo (sem lógica)
CSS   →  estilos, organizados por feature (espelha estrutura do JS)
JS    →  comportamento do browser (fetch, DOM, redirecionamentos)
PHP   →  sessão e proxy para API (nunca exposto como template)
```

### 4. Comunicação entre camadas

```
[Browser JS]
    │ fetch /php/class/Auth/auth.php
    ▼
[PHP front — auth.php]
    │ CurlHelper → http://nginx/api/auth/login
    ▼
[nginx]
    │ /api/* → back:8080 (sem o prefixo /api)
    ▼
[Go back — providers → rate_limit → user handler]
    │ usecase.Login → service.Authenticate → repository → GORM
    ▼
[MySQL]
    │ retorna linha da tabela users
    ▲
[Go back]
    │ response.Success({ token, expiresAt, user })
    ▲
[nginx → PHP front]
    │ $_SESSION['user'] = user; $_SESSION['token'] = token
    ▲
[Browser JS]
    │ window.location = '/'
```

## Fluxo de sessão

```
Sessão PHP vive no container front.
Cookie PHPSESSID persiste no browser.
JS nunca lê/grava a sessão diretamente.
JS chama session_check.php para saber o estado.
```

O backend é stateless: autentica com JWT (`Authorization: Bearer`). O front guarda o token na sessão PHP e o envia via `new CurlHelper($_SESSION['token'])` ao chamar rotas protegidas.

## Escalabilidade considerada no design

| Preocupação | Decisão tomada |
|---|---|
| Separar front/back | Containers independentes — escalam separadamente |
| JWT na API | Back stateless — escala horizontalmente. Sessão PHP do front: para várias instâncias, trocar por Redis session handler |
| Rate limit em memória | Vale por instância da API. Para várias réplicas, mover o contador para o Redis |
| Redis disponível | Pronto para uso como cache, rate-limiting ou session store distribuída |
| MySQL healthcheck + migrate | `migrate` só roda com MySQL saudável; `back` só sobe após o `migrate` terminar com sucesso |
| Volumes nomeados | Dados de MySQL e Redis não são perdidos em `docker compose down` |
