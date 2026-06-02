<?php

declare(strict_types=1);

namespace App\Core\Middleware;

/**
 * Registro central de middlewares da aplicação.
 *
 * Mapeia aliases (usados nas rotas) para as classes concretas.
 * Novos middlewares devem ser registrados aqui ou via register().
 *
 * Uso nas rotas:
 *   Route::get('/user', 'Domain\\User\\Controller\\UserController@index', ['auth']);
 */
class MiddlewareRegistry
{
    /** @var array<string, class-string<MiddlewareInterface>> */
    private static array $map = [
        'auth' => AuthMiddleware::class,
    ];

    /**
     * Resolve um alias para uma instância do middleware correspondente.
     *
     * @throws \RuntimeException se o alias não estiver registrado.
     */
    public static function resolve(string $alias): MiddlewareInterface
    {
        $class = self::$map[$alias] ?? null;

        if ($class === null) {
            throw new \RuntimeException(
                "Middleware '{$alias}' não encontrado no MiddlewareRegistry."
            );
        }

        return new $class();
    }

    /**
     * Registra um novo middleware em tempo de execução.
     *
     * @param class-string<MiddlewareInterface> $class
     */
    public static function register(string $alias, string $class): void
    {
        self::$map[$alias] = $class;
    }

    /**
     * Retorna todos os aliases registrados (útil para debug/testes).
     *
     * @return array<string, class-string<MiddlewareInterface>>
     */
    public static function all(): array
    {
        return self::$map;
    }
}
