<?php

declare(strict_types=1);

require_once __DIR__ . '/../helper/Session.php';

header('Content-Type: application/json; charset=utf-8');

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    header('Allow: POST');
    echo json_encode(['status' => false, 'message' => 'Método não permitido.']);
    exit;
}

Session::destroy();

echo json_encode(['status' => true]);
