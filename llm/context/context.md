# Contexto Geral do Projeto

## O que é
Template de projeto full-stack com arquitetura limpa, separação total de responsabilidades entre frontend e backend, pronto para Docker. Serve como ponto de partida para aplicações web com API em Go (DDD) e frontend sem framework.

## Repositório
```
simple-project-template/
├── back/      API Go (DDD, net/http + GORM)
├── front/     Frontend (HTML, CSS, JS, PHP de sessão)
├── nginx/     Configuração do proxy reverso
├── docker/    Dockerfile do front (o do back fica em back/docker/)
├── redis/     (volume Redis — não versionar dados)
└── llm/       Documentação para uso com IA
```

## Stack completa
| Camada | Tecnologia |
|---|---|
| Proxy reverso | nginx 1.25 |
| API | Go 1.22, net/http, GORM (MySQL), JWT |
| Frontend | HTML5, CSS3, JS ES2022, PHP 8.2-FPM (sessão + cURL) |
| Banco de dados | MySQL 8.0 |
| Cache / filas | Redis 7 |
| Orquestração | Docker Compose |

## Serviços Docker
| Serviço | Imagem | Porta interna | Porta exposta |
|---|---|---|---|
| nginx | nginx:1.25-alpine | 80 | `${NGINX_PORT:-80}` |
| migrate | build de back/docker/Dockerfile | — | — (roda e encerra) |
| back | build de back/docker/Dockerfile | 8080 | — |
| front | php:8.2-fpm-alpine | 9000 | — |
| mysql | mysql:8.0 | 3306 | `${DB_PORT_EXPOSE:-3306}` |
| redis | redis:7-alpine | 6379 | `${REDIS_PORT_EXPOSE:-6379}` |

Todos na rede interna `app_network`. Comunicação por nome de serviço (ex: `mysql`, `redis`, `nginx`).

## Arquivos de ambiente
| Arquivo | Quem usa |
|---|---|
| `.env` (raiz) | docker-compose — nomes de container, credenciais MySQL, portas expostas |
| `back/.env` | API Go (e serviço migrate) — banco, JWT, rate limit, CORS, mail, seed |
| `front/.env` | PHP do front — `API_BASE_URL`, `API_TIMEOUT` |

`back/.env` e `front/.env` são carregados nos containers via `env_file:` no docker-compose.

## Roteamento nginx
```
/api/*   → back container (Go :8080) — proxy_pass remove o prefixo /api
/*       → nginx serve estático de /var/www/front
/*.php   → front container (PHP-FPM :9000)
```

## Comunicação interna front → API
O PHP do container front chama a API através do nginx (na rede Docker):
```
front container → HTTP http://nginx/api/recurso → nginx → back container
```
Isso preserva o roteamento centralizado e evita acoplamento direto entre os containers de PHP.

## Padrão de autenticação
- Login com **login e senha**: a API valida e retorna `{ token (JWT), expiresAt, user }`
- O PHP do frontend guarda `user` e `token` na sessão (`$_SESSION`)
- Rotas protegidas da API exigem `Authorization: Bearer <token>`
- O JS verifica a sessão chamando `session_check.php`
- Páginas protegidas incluem `js/Auth/session-check.js` como primeiro script

## Como subir o projeto
```bash
cp .env.example .env
cp back/.env.example back/.env
cp front/.env.example front/.env
# edite os .env com suas credenciais
docker compose up -d --build
```

## Banco de dados
O schema é criado pelo executável `back/cmd/migration` (GORM AutoMigrate dos models listados em `back/migration/migrate.go`). O serviço `migrate` do docker-compose roda `migration -seed` antes da API subir, criando o usuário admin definido por `SEED_ADMIN_*` em `back/.env`.
