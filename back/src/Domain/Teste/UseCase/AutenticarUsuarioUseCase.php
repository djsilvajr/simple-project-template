<?php

declare(strict_types=1);

namespace App\Domain\Teste\UseCase;

use App\Domain\Teste\Repository\UsuarioRepositoryInterface;

class AutenticarUsuarioUseCase
{
    public function __construct(
        private UsuarioRepositoryInterface $repository
    ) {}

    public function executar(string $email, string $senha): array
    {
        if ($email === '' || $senha === '') {
            throw new \InvalidArgumentException('Email e senha são obrigatórios.');
        }

        $usuario = $this->repository->buscarPorEmail($email);

        if ($usuario === null) {
            throw new \RuntimeException('Credenciais inválidas.');
        }

        $pepper = $_ENV['PASSWORD_SECRET'] ?? '';

        if (!password_verify($senha . $pepper, $usuario['senha'])) {
            throw new \RuntimeException('Credenciais inválidas.');
        }

        unset($usuario['senha']);

        return $usuario;
    }
}
