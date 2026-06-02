<?php

declare(strict_types=1);

namespace App\Core;

use App\Core\Middleware\MiddlewareRegistry;

class Core
{
    public function dispatch(array $routes): void
    {
        $requestMethod = $_SERVER['REQUEST_METHOD'];
        $requestUri    = $_SERVER['REQUEST_URI'];

        $path = (string) parse_url($requestUri, PHP_URL_PATH);
        $path = $this->stripProjectPrefix($path);
        $path = '/' . trim($path, '/');

        foreach ($routes as $route) {
            if ($route['method'] !== $requestMethod) {
                continue;
            }

            $params = $this->matchRoute($route['path'], $path);

            if ($params === false) {
                continue;
            }

            // ── Pipeline de Middleware ────────────────────────────────────
            foreach ($route['middleware'] as $alias) {
                $middleware = MiddlewareRegistry::resolve($alias);
                $blocked    = $middleware->handle();

                if ($blocked !== null) {
                    header('Content-Type: application/json; charset=utf-8');
                    echo json_encode($blocked, JSON_UNESCAPED_UNICODE);
                    return;
                }
            }
            // ─────────────────────────────────────────────────────────────

            [$controllerName, $action] = explode('@', $route['action'], 2);

            $controllerClass = $this->resolveControllerClass($controllerName);

            if (!class_exists($controllerClass)) {
                $this->respond(404, "Controller '{$controllerName}' não encontrado.");
                return;
            }

            $controllerInstance = new $controllerClass();

            if (!method_exists($controllerInstance, $action)) {
                $this->respond(404, "Método '{$action}' não encontrado no controller.");
                return;
            }

            header('Content-Type: application/json; charset=utf-8');
            $result = $controllerInstance->$action($params);

            if (is_array($result)) {
                echo json_encode($result, JSON_UNESCAPED_UNICODE);
            }

            return;
        }

        $this->respond(404, "Rota '{$path}' não encontrada.");
    }

    /**
     * Verifica se o path da requisição bate com o padrão da rota.
     * Suporta parâmetros dinâmicos no formato {param}.
     *
     * @return array<int, string>|false
     */
    private function matchRoute(string $routePath, string $requestPath): array|false
    {
        $pattern = preg_replace('/\{[^}]+\}/', '([^/]+)', $routePath);
        $pattern = '#^' . $pattern . '$#';

        if (preg_match($pattern, $requestPath, $matches) === 1) {
            array_shift($matches);
            return $matches;
        }

        return false;
    }

    /**
     * Resolve o FQCN do controller.
     * Aceita namespace completo (Domain\...) ou nome simples legado (TestController).
     */
    private function resolveControllerClass(string $controllerName): string
    {
        if (str_contains($controllerName, '\\')) {
            return 'App\\' . ltrim($controllerName, '\\');
        }

        return 'App\\Controllers\\' . $controllerName;
    }

    /**
     * Remove o prefixo do PROJECT_NAME do path quando rodando em subpasta.
     * Ex.: localhost/php-api-skeleton/user → /user
     */
    private function stripProjectPrefix(string $path): string
    {
        $projectName = $_ENV['PROJECT_NAME'] ?? '';

        if ($projectName === '') {
            return $path;
        }

        $prefix = '/' . ltrim($projectName, '/');

        if (str_starts_with($path, $prefix)) {
            return substr($path, strlen($prefix)) ?: '/';
        }

        return $path;
    }

    private function respond(int $statusCode, string $message): void
    {
        http_response_code($statusCode);
        header('Content-Type: application/json; charset=utf-8');
        echo json_encode([
            'status' => false,
            'error'  => $message,
        ], JSON_UNESCAPED_UNICODE);
    }
}
