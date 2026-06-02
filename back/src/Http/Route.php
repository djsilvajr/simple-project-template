<?php

declare(strict_types=1);

namespace App\Http;

class Route
{
    /** @var array<int, array{path: string, action: string, method: string, middleware: string[]}> */
    private static array $routes = [];

    /**
     * @param string[] $middleware  Aliases de middleware a executar antes do controller.
     */
    public static function get(string $path, string $action, array $middleware = []): void
    {
        self::register('GET', $path, $action, $middleware);
    }

    /**
     * @param string[] $middleware
     */
    public static function post(string $path, string $action, array $middleware = []): void
    {
        self::register('POST', $path, $action, $middleware);
    }

    /**
     * @param string[] $middleware
     */
    public static function put(string $path, string $action, array $middleware = []): void
    {
        self::register('PUT', $path, $action, $middleware);
    }

    /**
     * @param string[] $middleware
     */
    public static function delete(string $path, string $action, array $middleware = []): void
    {
        self::register('DELETE', $path, $action, $middleware);
    }

    /**
     * @return array<int, array{path: string, action: string, method: string, middleware: string[]}>
     */
    public static function routes(): array
    {
        return self::$routes;
    }

    /**
     * @param string[] $middleware
     */
    private static function register(
        string $method,
        string $path,
        string $action,
        array  $middleware,
    ): void {
        self::$routes[] = [
            'path'       => '/' . ltrim($path, '/'),
            'action'     => $action,
            'method'     => strtoupper($method),
            'middleware' => $middleware,
        ];
    }
}
