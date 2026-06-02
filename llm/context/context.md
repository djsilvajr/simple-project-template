# Contexto Geral do Projeto

## O que é
Template de projeto PHP full-stack com arquitetura limpa, separação total de responsabilidades entre frontend e backend, pronto para Docker. Serve como ponto de partida para aplicações web com API PHP pura e frontend sem framework.

## Repositório
```
simple-project-template/
├── back/      API PHP (DDD, sem framework)
├── front/     Frontend (HTML, CSS, JS, PHP de sessão)
├── nginx/     Configuração do proxy reverso
├── docker/    Dockerfiles de back e front
├── db/        Scripts SQL de inicialização
├── redis/     (volume Redis — não versionar dados)
└── llm/       Documentação para uso com IA
```

## Stack completa
| Camada | Tecnologia |
|---|---|
| Proxy reverso | nginx 1.25 |
| API | PHP 8.2-FPM, PDO MySQL, phpdotenv v5 |
| Frontend | HTML5, CSS3, JS ES2022, PHP 8.2-FPM (sessão + cURL) |
| Banco de dados | MySQL 8.0 |
| Cache / filas | Redis 7 |
| Orquestração | Docker Compose |

## Serviços Docker
| Serviço | Imagem | Porta interna | Porta exposta |
|---|---|---|---|
| nginx | nginx:1.25-alpine | 80 | `${NGINX_PORT:-80}` |
| back | php:8.2-fpm-alpine | 9000 | — |
| front | php:8.2-fpm-alpine | 9000 | — |
| mysql | mysql:8.0 | 3306 | `${DB_PORT_EXPOSE:-3306}` |
| redis | redis:7-alpine | 6379 | `${REDIS_PORT_EXPOSE:-6379}` |

Todos na rede interna `app_network`. Comunicação por nome de serviço (ex: `mysql`, `redis`, `nginx`).

## Arquivos de ambiente
| Arquivo | Quem usa |
|---|---|
| `.env` (raiz) | docker-compose — nomes de container, credenciais MySQL, portas expostas |
| `back/.env` | PHP backend — DB_HOST, PROJECT_NAME, PASSWORD_SECRET, etc. |

O frontend não tem `.env` próprio. `API_BASE_URL=http://nginx/api` é injetado via `environment:` no docker-compose do serviço `front`.

## Roteamento nginx
```
/api/*   → back container (PHP-FPM :9000) — roda index.php para toda requisição
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
- A API valida credenciais e retorna dados do usuário (sem JWT)
- O PHP do frontend cria e mantém a sessão (`$_SESSION`)
- O JS verifica a sessão chamando `session_check.php`
- Páginas protegidas incluem `js/Auth/session-check.js` como primeiro script

## Como subir o projeto
```bash
cp .env.example .env
cp back/.env.exemple back/.env
# edite os .env com suas credenciais
docker compose up -d --build
```

## Banco de dados
Scripts SQL em `db/` são executados automaticamente pelo MySQL na primeira inicialização do container (montado em `/docker-entrypoint-initdb.d`).
