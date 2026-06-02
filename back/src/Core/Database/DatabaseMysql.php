<?php

namespace App\Core\Database;

use PDO;
use PDOException;

class DatabaseMysql
{
    /**
     * Instância do PDO
     */
    private PDO $pdo;

    public function __construct()
    {
        $host     = $_ENV['DB_HOST']     ?? getenv('DB_HOST')     ?: '127.0.0.1';
        $port     = $_ENV['DB_PORT']     ?? getenv('DB_PORT')     ?: '3306';
        $database = $_ENV['DB_DATABASE'] ?? getenv('DB_DATABASE') ?: '';
        $username = $_ENV['DB_USERNAME'] ?? getenv('DB_USERNAME') ?: '';
        $password = $_ENV['DB_PASSWORD'] ?? getenv('DB_PASSWORD') ?: '';

        // DSN MySQL com charset
        $dsn = "mysql:host={$host};port={$port};dbname={$database};charset=utf8mb4";

        // Options recomendadas
        $options = [
            PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
            PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
            PDO::ATTR_EMULATE_PREPARES   => false, // usa prepared statements nativos quando disponível
            // PDO::ATTR_PERSISTENT       => true, // opcional: habilitar conexões persistentes com cautela
            PDO::MYSQL_ATTR_INIT_COMMAND => "SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci",
        ];

        try {
            $this->pdo = new PDO($dsn, $username, $password, $options);
        } catch (PDOException $e) {
            // Em produção prefira log e uma resposta adequada em vez de die()
            throw new PDOException('Erro de conexão com MySQL: ' . $e->getMessage(), (int)$e->getCode());
        }
    }

    /**
     * SELECT com múltiplos resultados
     */
    public function selectAll(string $query, array $params = []): array
    {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
        return $stmt->fetchAll(PDO::FETCH_ASSOC);
    }

    /**
     * SELECT com um único resultado
     */
    public function selectOne(string $query, array $params = []): ?array
    {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        return $result ?: null;
    }

    /**
     * UPDATE genérico
     */
    public function update(string $query, array $params = []): int
    {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
        return $stmt->rowCount();
    }

    /**
     * DELETE genérico
     */
    public function delete(string $query, array $params = []): int
    {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
        return $stmt->rowCount();
    }

    /**
     * INSERT genérico
     */
    public function insert(string $query, array $params = []): int
    {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
        return (int) $this->pdo->lastInsertId();
    }

    public function insertNoKeyTable(string $query, array $params = []): int {
        $stmt = $this->pdo->prepare($query);
        $stmt->execute($params);
       return $stmt->rowCount();
    }
}
