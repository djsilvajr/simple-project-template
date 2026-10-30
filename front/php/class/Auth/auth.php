<?php

declare(strict_types=1);

require_once __DIR__ . '/../helper/Session.php';

Session::start();

header('Content-Type: application/json; charset=utf-8');

$input = json_decode(file_get_contents('php://input'), true) ?? [];
$login    = trim($input['login'] ?? '');
$password = $input['password'] ?? '';

if ($login === '' || $password === '') {
    http_response_code(422);
    echo json_encode(['status' => false, 'message' => 'Login e senha são obrigatórios.']);
    exit;
}

$curl     = new CurlHelper();
$response = $curl->post('/auth/login', ['login' => $login, 'password' => $password]);

if ($response['status'] === 200 && !empty($response['body']['status'])) {
    session_regenerate_id(true);

    Session::storeAuth($response['body']['data']);  // user + JWT para chamar rotas protegidas da API

    echo json_encode(['status' => true]);
    exit;
}

http_response_code(401);
$message = $response['body']['error'] ?? 'Credenciais inválidas.';
echo json_encode(['status' => false, 'message' => $message]);
