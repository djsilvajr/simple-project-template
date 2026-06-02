# Arquitetura do Backend

## Visão geral das camadas

```
┌─────────────────────────────────────────────────┐
│  nginx                                          │
│  Recebe /api/* e encaminha para PHP-FPM :9000   │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  index.php  (bootstrap)                         │
│  1. Carrega back/.env via phpdotenv             │
│  2. Inclui Routes/main.php (registra rotas)     │
│  3. Instancia Core e chama dispatch()           │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  Core/Core.php  (dispatcher)                    │
│  1. Lê REQUEST_METHOD + REQUEST_URI             │
│  2. Remove prefixo PROJECT_NAME do path         │
│  3. Itera rotas: compara método + path regex    │
│  4. Executa pipeline de middleware              │
│  5. Instancia controller e chama action($params)│
│  6. Serializa retorno como JSON                 │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  Domain / Controller                            │
│  - Lê body JSON do request                     │
│  - Instancia UseCase + Repository               │
│  - Retorna Response::success() ou error()       │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  Domain / UseCase                               │
│  - Validações de negócio                        │
│  - Orquestra chamadas ao Repository             │
│  - Lança exceções tipadas (\InvalidArgumentException, \RuntimeException)
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  Domain / Repository                            │
│  - Implementa RepositoryInterface               │
│  - Único ponto de acesso ao banco               │
│  - Injeta DatabaseMysql                         │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  Core/Database/DatabaseMysql                    │
│  - PDO com prepared statements nativos          │
│  - UTF-8mb4, ERRMODE_EXCEPTION                  │
│  - Métodos: selectAll, selectOne, insert,       │
│    insertNoKeyTable, update, delete             │
└─────────────────┬───────────────────────────────┘
                  │
              MySQL 8.0
```

## Estrutura de pastas

```
back/
├── index.php                  bootstrap
├── composer.json              PSR-4: App\ → src/
├── src/
│   ├── Core/
│   │   ├── Core.php           dispatcher de rotas
│   │   ├── Response.php       helper de resposta JSON padronizada
│   │   ├── Database/
│   │   │   └── DatabaseMysql.php   abstração PDO
│   │   └── Middleware/
│   │       ├── MiddlewareInterface.php
│   │       ├── MiddlewareRegistry.php   mapa alias → classe
│   │       ├── AuthMiddleware.php       HTTP Basic Auth
│   │       └── Response.php
│   ├── Http/
│   │   └── Route.php          API fluente de registro de rotas
│   ├── Routes/
│   │   └── main.php           ponto central de definição de rotas
│   └── Domain/
│       └── <NomeDominio>/     um diretório por bounded context
│           ├── Controller/
│           ├── UseCase/
│           ├── Repository/
│           └── DTO/
```

## Padrão de rota → controller

```
Route::post('/usuarios', 'Domain\\Teste\\Controller\\UsuarioController@criar', []);
                                                                               ↑
                                                            array de aliases de middleware
```

O `Core.php` resolve `App\Domain\Teste\Controller\UsuarioController` via PSR-4 e chama `criar($params)`.

## Middleware pipeline

```
foreach ($route['middleware'] as $alias) {
    $resultado = MiddlewareRegistry::resolve($alias)->handle();
    if ($resultado !== null) { echo json_encode($resultado); return; }
}
// só chega aqui se todos os middlewares retornaram null
controller->action($params);
```

## Ciclo de vida de um erro

| Tipo de exceção | HTTP sugerido | Lançado em |
|---|---|---|
| `\InvalidArgumentException` | 422 | UseCase — validação de entrada |
| `\RuntimeException` | 409 / 401 | UseCase — regra de negócio / auth |
| `\Throwable` | 500 | Controller — catch-all |
