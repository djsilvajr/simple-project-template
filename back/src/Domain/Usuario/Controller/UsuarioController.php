<?php

declare(strict_types=1);

namespace App\Domain\Usuario\Controller;

use App\Core\Response;
use App\Domain\Usuario\DTO\UsuarioDTO;
use App\Domain\Usuario\Repository\UsuarioRepository;
use App\Domain\Usuario\UseCase\CriarUsuarioUseCase;

class UsuarioController
{
    public function criar(array $params): array
    {
        $body = json_decode(file_get_contents('php://input'), true) ?? [];

        try {
            $dto     = UsuarioDTO::fromArray($body);
            $useCase = new CriarUsuarioUseCase(new UsuarioRepository());
            $usuario = $useCase->executar($dto);

            return Response::success($usuario, 201);
        } catch (\InvalidArgumentException $e) {
            return Response::error($e->getMessage(), 422);
        } catch (\RuntimeException $e) {
            return Response::error($e->getMessage(), 409);
        } catch (\Throwable $e) {
            return Response::error('Erro interno no servidor.', 500);
        }
    }
}
