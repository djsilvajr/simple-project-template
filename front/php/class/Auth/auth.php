<?php

declare(strict_types=1);

session_start();

header('Content-Type: application/json; charset=utf-8');

require_once __DIR__ . '/../helper/CurlHelper.php';

$input = json_decode(file_get_contents('php://input'), true) ?? [];
$email = trim($input['email'] ?? '');
$senha = $input['senha'] ?? '';

if ($email === '' || $senha === '') {
    http_response_code(422);
    echo json_encode(['status' => false, 'message' => 'Email e senha são obrigatórios.']);
    exit;
}

$curl     = new CurlHelper();
$response = $curl->post('/auth/login', ['email' => $email, 'senha' => $senha]);

if ($response['status'] === 200 && !empty($response['body']['status'])) {
    $_SESSION['user'] = $response['body']['data'];
    echo json_encode(['status' => true]);
    exit;
}

http_response_code(401);
$message = $response['body']['error'] ?? 'Credenciais inválidas.';
echo json_encode(['status' => false, 'message' => $message]);
