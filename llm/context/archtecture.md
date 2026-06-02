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
                        │  /api/*  → back:9000             │
                        │  /*.php  → front:9000            │
                        │  /*      → static /var/www/front │
                        └────────┬──────────────┬──────────┘
                                 │              │
              ┌──────────────────▼──┐    ┌──────▼──────────────────┐
              │  back (PHP-FPM)     │    │  front (PHP-FPM)        │
              │  :9000              │    │  :9000                  │
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
- **back**: PHP-FPM sem servidor HTTP próprio. Só processa o que nginx encaminha.
- **front**: PHP-FPM para arquivos `.php` do frontend. Arquivos estáticos são servidos pelo nginx diretamente.
- **mysql**: banco relacional. Inicializado com scripts de `db/`.
- **redis**: disponível para cache, filas ou sessões distribuídas.

### 2. Backend (back/)
Arquitetura DDD (Domain-Driven Design):
```
Core         →  infraestrutura: roteador, PDO, middleware, response
Http         →  registro de rotas (API fluente)
Routes       →  ponto central de declaração de rotas
Domain       →  lógica de negócio por bounded context
  Controller →  entrada HTTP: lê request, retorna Response
  UseCase    →  regras de negócio puras, sem I/O direto
  Repository →  abstração de acesso a dados
  DTO        →  objetos de transporte entre camadas
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
    │ /api/* → back:9000
    ▼
[PHP back — AuthController]
    │ AutenticarUsuarioUseCase → UsuarioRepository → DatabaseMysql
    ▼
[MySQL]
    │ retorna linha da tabela usuarios
    ▲
[PHP back]
    │ Response::success(usuario_sem_senha)
    ▲
[nginx → PHP front]
    │ $_SESSION['user'] = dados_usuario
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

Isso significa: sessão não é compartilhada com o backend. O backend é stateless — apenas valida credenciais. O estado de autenticação fica no frontend PHP.

## Escalabilidade considerada no design

| Preocupação | Decisão tomada |
|---|---|
| Separar front/back | Containers independentes — escalam separadamente |
| Sem JWT no template | Sessão PHP é suficiente para single-instance. Para múltiplas instâncias de front, trocar por Redis session handler |
| Redis disponível | Pronto para uso como cache, rate-limiting ou session store distribuída |
| MySQL healthcheck | Back só sobe após MySQL estar respondendo |
| Volumes nomeados | Dados de MySQL e Redis não são perdidos em `docker compose down` |
