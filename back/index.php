<?php

declare(strict_types=1);

require_once __DIR__ . '/vendor/autoload.php';

// Carrega variáveis de ambiente antes de qualquer outra coisa
$dotenv = Dotenv\Dotenv::createImmutable(__DIR__);
$dotenv->load();

// Registra as rotas da aplicação
require_once __DIR__ . '/src/Routes/main.php';

use App\Core\Core;
use App\Http\Route;

$routes = Route::routes();
$core   = new Core();
$core->dispatch($routes);
