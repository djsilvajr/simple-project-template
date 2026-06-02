<?php

declare(strict_types=1);

namespace App\Domain\Teste\Repository;

use App\Core\Database\DatabaseMysql;

class UsuarioRepository implements UsuarioRepositoryInterface
{
    private DatabaseMysql $db;

    public function __construct()
    {
        $this->db = new DatabaseMysql();
    }

    public function criar(string $usuario, string $email, string $senhaHash): int
    {
        return $this->db->insert(
            'INSERT INTO usuarios (usuario, email, senha) VALUES (:usuario, :email, :senha)',
            ['usuario' => $usuario, 'email' => $email, 'senha' => $senhaHash]
        );
    }

    public function buscarPorEmail(string $email): ?array
    {
        return $this->db->selectOne(
            'SELECT * FROM usuarios WHERE email = :email LIMIT 1',
            ['email' => $email]
        );
    }

    public function emailExiste(string $email): bool
    {
        return $this->db->selectOne(
            'SELECT 1 FROM usuarios WHERE email = :email LIMIT 1',
            ['email' => $email]
        ) !== null;
    }
}
