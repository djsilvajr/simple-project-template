<?php

declare(strict_types=1);

namespace App\Domain\Usuario\DTO;

class UsuarioDTO
{
    public function __construct(
        public readonly string $usuario,
        public readonly string $email,
        public readonly string $senha
    ) {}

    public static function fromArray(array $data): self
    {
        return new self(
            usuario: trim($data['usuario'] ?? ''),
            email:   trim($data['email']   ?? ''),
            senha:   $data['senha']        ?? ''
        );
    }
}
