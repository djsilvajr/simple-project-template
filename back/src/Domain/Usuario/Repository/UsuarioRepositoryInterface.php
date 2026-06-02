<?php

declare(strict_types=1);

namespace App\Domain\Usuario\Repository;

interface UsuarioRepositoryInterface
{
    public function criar(string $usuario, string $email, string $senhaHash): int;
    public function buscarPorEmail(string $email): ?array;
    public function emailExiste(string $email): bool;
}
