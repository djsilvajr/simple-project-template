<?php

declare(strict_types=1);

namespace App\Core;

/**
 * Helper para construir respostas JSON padronizadas.
 *
 * Uso:
 *   return Response::success($user->toArray());
 *   return Response::success($user->toArray(), 201);
 *   return Response::error("Não encontrado.", 404);
 */
class Response
{
    /**
     * Resposta de sucesso (2xx).
     *
     * @param mixed $data
     */
    public static function success(mixed $data, int $statusCode = 200): array
    {
        http_response_code($statusCode);

        return [
            'status'     => true,
            'statusCode' => $statusCode,
            'data'       => $data,
        ];
    }

    /**
     * Resposta de erro (4xx / 5xx).
     */
    public static function error(string $message, int $statusCode = 400): array
    {
        http_response_code($statusCode);

        return [
            'status'     => false,
            'statusCode' => $statusCode,
            'error'      => $message,
        ];
    }
}
