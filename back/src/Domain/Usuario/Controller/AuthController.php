<?php

declare(strict_types=1);

namespace App\Domain\Usuario\Controller;

use App\Core\Response;
use App\Domain\Usuario\Repository\UsuarioRepository;
use App\Domain\Usuario\UseCase\AutenticarUsuarioUseCase;

class AuthController
{
    public function login(array $params): array
    {
        $body  = json_decode(file_get_contents('php://input'), true) ?? [];
        $email = trim($body['email'] ?? '');
        $senha = $body['senha'] ?? '';

        try {
            $useCase = new AutenticarUsuarioUseCase(new UsuarioRepository());
            $usuario = $useCase->executar($email, $senha);

            return Response::success($usuario);
        } catch (\InvalidArgumentException $e) {
            return Response::error($e->getMessage(), 422);
        } catch (\RuntimeException $e) {
            return Response::error($e->getMessage(), 401);
        } catch (\Throwable $e) {
            return Response::error('Erro interno no servidor.', 500);
        }
    }
}
