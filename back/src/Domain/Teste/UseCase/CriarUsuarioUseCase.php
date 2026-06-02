<?php

declare(strict_types=1);

namespace App\Domain\Teste\UseCase;

use App\Domain\Teste\DTO\UsuarioDTO;
use App\Domain\Teste\Repository\UsuarioRepositoryInterface;

class CriarUsuarioUseCase
{
    public function __construct(
        private UsuarioRepositoryInterface $repository
    ) {}

    public function executar(UsuarioDTO $dto): array
    {
        if ($dto->usuario === '') {
            throw new \InvalidArgumentException('Campo usuario é obrigatório.');
        }

        if ($dto->email === '' || !filter_var($dto->email, FILTER_VALIDATE_EMAIL)) {
            throw new \InvalidArgumentException('Email inválido.');
        }

        if (strlen($dto->senha) < 6) {
            throw new \InvalidArgumentException('Senha deve ter no mínimo 6 caracteres.');
        }

        if ($this->repository->emailExiste($dto->email)) {
            throw new \RuntimeException('Email já cadastrado.');
        }

        $pepper    = $_ENV['PASSWORD_SECRET'] ?? '';
        $senhaHash = password_hash($dto->senha . $pepper, PASSWORD_BCRYPT);

        $id = $this->repository->criar($dto->usuario, $dto->email, $senhaHash);

        return [
            'id'               => $id,
            'usuario'          => $dto->usuario,
            'email'            => $dto->email,
            'criacao_datahora' => date('Y-m-d H:i:s'),
        ];
    }
}
