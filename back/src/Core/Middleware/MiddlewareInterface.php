<?php

declare(strict_types=1);

namespace App\Core\Middleware;

/**
 * Contrato para todos os middlewares da aplicação.
 *
 * O método handle() deve:
 *   - Retornar null para permitir que a requisição continue.
 *   - Retornar um array de resposta para interromper o pipeline
 *     (o Core irá serializar e enviar esse array como JSON).
 */
interface MiddlewareInterface
{
    /**
     * Processa a requisição antes de chegar ao controller.
     *
     * @return array<string, mixed>|null  null = continuar | array = resposta de bloqueio
     */
    public function handle(): ?array;
}
