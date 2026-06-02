# Contexto do Backend

## Stack
- PHP 8.2 puro — sem framework (sem Laravel, sem Symfony)
- MySQL 8.0 via PDO (`DatabaseMysql`)
- Redis 7 (disponível via serviço `redis` na rede Docker)
- `vlucas/phpdotenv` v5 para carregar variáveis de ambiente
- Composer com PSR-4: `App\` → `src/`

## Variáveis de ambiente (`back/.env`)
| Variável | Uso |
|---|---|
| `PROJECT_NAME` | Prefixo de URL removido pelo roteador — deve ser `api` |
| `DB_HOST` | Host MySQL — use `mysql` (nome do serviço Docker) |
| `DB_PORT` | Porta MySQL (padrão `3306`) |
| `DB_DATABASE` | Nome do banco |
| `DB_USERNAME` | Usuário MySQL |
| `DB_PASSWORD` | Senha MySQL |
| `API_USER` | Usuário para HTTP Basic Auth (middleware `auth`) |
| `API_PASSWORD` | Senha para HTTP Basic Auth |
| `PASSWORD_SECRET` | Pepper concatenado à senha antes do hash/verify |
| `REDIS_HOST` | Host Redis — use `redis` (nome do serviço Docker) |
| `REDIS_PORT` | Porta Redis (padrão `6379`) |

## Fluxo de uma requisição
```
nginx /api/recurso → PHP-FPM back:9000
  → index.php            carrega .env, registra rotas
    → Routes/main.php    define Route::post/get/put/delete
      → Core.php         match de URL, executa middleware, chama controller
        → Controller     lê body, chama UseCase, retorna array via Response::*
          → UseCase      regras de negócio, valida dados, chama Repository
            → Repository → DatabaseMysql → MySQL
```

> `PROJECT_NAME=api` faz `Core::stripProjectPrefix` remover `/api` antes do match.
> Registre rotas **sem** o prefixo `/api`: `Route::post('/usuarios', ...)`.

## Padrão de resposta
```php
use App\Core\Response;

return Response::success($dados);         // HTTP 200
return Response::success($dados, 201);    // HTTP 201 Created
return Response::error('Mensagem', 422);  // HTTP 4xx / 5xx
```
`Response` já define `http_response_code` e retorna o array que `Core.php` serializa como JSON.

## Como adicionar um novo domínio
1. Crie `src/Domain/NomeDominio/` com subpastas `Controller/`, `UseCase/`, `Repository/`, `DTO/`
2. Implemente `RepositoryInterface` e `Repository` (injeta `DatabaseMysql`)
3. Crie `UseCase` com regras de negócio
4. Crie `Controller` — lê `json_decode(file_get_contents('php://input'), true)`, chama use case, retorna `Response::*`
5. Registre as rotas em `src/Routes/main.php`

Referência: `Domain/Teste` é o exemplo completo funcional.

## DatabaseMysql — métodos disponíveis
```php
$db->selectAll($sql, $params);      // array de linhas
$db->selectOne($sql, $params);      // array|null (uma linha)
$db->insert($sql, $params);         // int (lastInsertId)
$db->insertNoKeyTable($sql, $params); // int (rowCount)
$db->update($sql, $params);         // int (rowCount)
$db->delete($sql, $params);         // int (rowCount)
```
Parâmetros sempre como array associativo com placeholders nomeados (`:coluna`).

## Hashing de senha com pepper
```php
$pepper = $_ENV['PASSWORD_SECRET'] ?? '';
$hash   = password_hash($senha . $pepper, PASSWORD_BCRYPT);  // criar
$ok     = password_verify($senha . $pepper, $hashDoBanco);   // verificar
```

## Middleware
- Implementar `App\Core\Middleware\MiddlewareInterface::handle(): ?array`
- `null` = continua para o controller; `array` = bloqueia com aquele array como resposta JSON
- Registrar alias em `MiddlewareRegistry::$map`
- Usar nas rotas: `Route::get('/rota', 'Controller@metodo', ['alias'])`

## Convenções
- Body JSON: `json_decode(file_get_contents('php://input'), true) ?? []`
- `$params` no controller = segmentos dinâmicos da URL (`{id}` etc.)
- `Content-Type: application/json` já é definido por `Core.php` antes de chamar o controller
- Exceções de validação → HTTP 422; conflito/negócio → HTTP 409; auth → HTTP 401
