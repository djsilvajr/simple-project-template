<?php

declare(strict_types=1);

/**
 * Ponto único de controle da sessão do front.
 *
 * Use Session::start() no lugar de session_start() em todo arquivo PHP que usa sessão.
 *
 * - Cookie com HttpOnly, SameSite=Lax e Secure (quando HTTPS ou SESSION_SECURE_COOKIE=true)
 * - session.gc_maxlifetime = SESSION_LIFETIME + SESSION_REFRESH_GRACE (minutos)
 * - Token da API expirado: tenta renovar em POST /auth/refresh se expirou há no máximo
 *   SESSION_REFRESH_GRACE minutos; se não der, a sessão é limpa (usuário volta ao login)
 */
require_once __DIR__ . '/CurlHelper.php';

final class Session
{
    public static function start(): void
    {
        if (session_status() === PHP_SESSION_ACTIVE) {
            return;
        }

        ini_set('session.gc_maxlifetime', (string) self::lifetimeSeconds());
        ini_set('session.use_strict_mode', '1');   // rejeita IDs de sessão não emitidos pelo servidor
        ini_set('session.use_only_cookies', '1');  // nunca aceita ID de sessão pela URL

        session_set_cookie_params(self::cookieParams());
        session_start();

        // Token da API expirado → tenta renovar; se não conseguir, descarta a sessão
        if (isset($_SESSION['token_expires_at']) && $_SESSION['token_expires_at'] <= time()) {
            if (!self::renewToken()) {
                $_SESSION = [];
                session_regenerate_id(true);
            }
        }
    }

    /**
     * Guarda na sessão o retorno de /auth/login ou /auth/refresh ({ token, expiresAt, user }).
     */
    public static function storeAuth(array $data): void
    {
        $_SESSION['user']             = $data['user'];
        $_SESSION['token']            = $data['token'];
        $_SESSION['token_expires_at'] = $data['expiresAt'];
    }

    /**
     * Troca o token expirado por um novo na API. Só tenta dentro da janela de
     * SESSION_REFRESH_GRACE minutos após a expiração (a API aplica a mesma regra).
     */
    private static function renewToken(): bool
    {
        $token = $_SESSION['token'] ?? '';
        if ($token === '' || time() - (int) $_SESSION['token_expires_at'] > self::refreshGraceSeconds()) {
            return false;
        }

        $response = (new CurlHelper($token))->post('/auth/refresh', []);

        if ($response['status'] !== 200 || empty($response['body']['status'])) {
            return false;
        }

        self::storeAuth($response['body']['data']);
        return true;
    }

    /**
     * Encerra a sessão por completo: dados, arquivo no servidor e cookie no navegador.
     */
    public static function destroy(): void
    {
        self::start();

        $_SESSION = [];

        $params = self::cookieParams();
        unset($params['lifetime']);
        setcookie(session_name(), '', ['expires' => time() - 3600] + $params);

        session_destroy();
    }

    /**
     * Inclui a janela de renovação, para o PHP não apagar a sessão (e o token)
     * antes de dar tempo de renovar.
     */
    private static function lifetimeSeconds(): int
    {
        return max(1, (int) (getenv('SESSION_LIFETIME') ?: 60)) * 60 + self::refreshGraceSeconds();
    }

    private static function refreshGraceSeconds(): int
    {
        $grace = getenv('SESSION_REFRESH_GRACE');
        return max(0, $grace === false || $grace === '' ? 30 : (int) $grace) * 60;
    }

    private static function cookieParams(): array
    {
        return [
            'lifetime' => 0,        // cookie de sessão do navegador; a validade real é controlada pelo token
            'path'     => '/',
            'secure'   => self::isSecure(),
            'httponly' => true,
            'samesite' => 'Lax',
        ];
    }

    private static function isSecure(): bool
    {
        if (filter_var(getenv('SESSION_SECURE_COOKIE'), FILTER_VALIDATE_BOOLEAN)) {
            return true;
        }

        return !empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off';
    }
}
