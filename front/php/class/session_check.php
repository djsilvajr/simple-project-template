<?php

declare(strict_types=1);

require_once __DIR__ . '/helper/Session.php';

// Session::start() já descarta a sessão se o token da API tiver expirado
Session::start();

header('Content-Type: application/json; charset=utf-8');

echo json_encode([
    'autenticado' => isset($_SESSION['user']),
    'usuario'     => $_SESSION['user'] ?? null,
]);
