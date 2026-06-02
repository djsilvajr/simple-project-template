<?php

declare(strict_types=1);

namespace App\Core\Middleware;

/**
 * Middleware de autenticação via HTTP Basic Auth.
 *
 * Verifica o header "Authorization: Basic <base64(user:senha)>" e compara
 * com as credenciais definidas no .env (API_USER e API_PASSWORD).
 *
 * Uso nas rotas:
 *   Route::get('/user', 'Domain\\User\\Controller\\UserController@index', ['auth']);
 *
 * Exemplo de requisição autenticada:
 *   curl -u admin:secret http://localhost/user
 *   curl -H "Authorization: Basic YWRtaW46c2VjcmV0" http://localhost/user
 */
class AuthMiddleware implements MiddlewareInterface
{
    public function handle(): ?array
    {
        $authHeader = $this->getAuthorizationHeader();

        if ($authHeader === null || !str_starts_with($authHeader, 'Basic ')) {
            return $this->unauthorized('Header de autorização ausente ou formato inválido.');
        }

        $decoded = base64_decode(substr($authHeader, 6), strict: true);

        if ($decoded === false || !str_contains($decoded, ':')) {
            return $this->unauthorized('Credenciais mal formatadas.');
        }

        [$username, $password] = explode(':', $decoded, 2);

        $validUser     = $_ENV['API_USER']     ?? '';
        $validPassword = $_ENV['API_PASSWORD'] ?? '';

        if ($validUser === '' || $validPassword === '') {
            return $this->unauthorized('Credenciais da API não configuradas no servidor.');
        }

        // Comparação segura contra timing attacks
        $userMatch = hash_equals($validUser, $username);
        $passMatch = hash_equals($validPassword, $password);

        if (!$userMatch || !$passMatch) {
            return $this->unauthorized('Usuário ou senha inválidos.');
        }

        return null; // Autenticado — continua para o controller
    }

    /**
     * Tenta obter o header Authorization em diferentes contextos de servidor
     * (Apache com mod_rewrite, Nginx, CGI/FastCGI, etc.).
     */
    private function getAuthorizationHeader(): ?string
    {
        if (isset($_SERVER['HTTP_AUTHORIZATION'])) {
            return $_SERVER['HTTP_AUTHORIZATION'];
        }

        // Apache com AllowOverride + mod_rewrite pode mover o header aqui
        if (isset($_SERVER['REDIRECT_HTTP_AUTHORIZATION'])) {
            return $_SERVER['REDIRECT_HTTP_AUTHORIZATION'];
        }

        // Fallback via apache_request_headers() quando disponível
        if (function_exists('apache_request_headers')) {
            $headers = apache_request_headers();
            if (isset($headers['Authorization'])) {
                return $headers['Authorization'];
            }
        }

        return null;
    }

    private function unauthorized(string $reason): array
    {
        http_response_code(401);
        header('WWW-Authenticate: Basic realm="PHP API"');

        return [
            'status'     => false,
            'statusCode' => 401,
            'error'      => 'Não autorizado.',
            'detail'     => $reason,
        ];
    }
}
