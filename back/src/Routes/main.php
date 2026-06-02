<?php

declare(strict_types=1);

use App\Http\Route;

/*
|--------------------------------------------------------------------------
| Rotas da Aplicação
|--------------------------------------------------------------------------
|
| Formato:
|   Route::METHOD('path', 'Namespace\Controller@metodo', ['middleware']);
|   EX : Route::get('/user','Domain\\User\\Controller\\UserController@index', ['auth']);
|
| Middleware disponíveis:
|   'auth'  →  HTTP Basic Auth (API_USER / API_PASSWORD no .env)
|
| Para controllers dentro de Domain, use o namespace a partir de App\:
|   'Domain\User\Controller\UserController@index'
|
| Para controllers legados em src/Controllers/, basta o nome simples:
|   'TestController@get'
|
*/

// --- Domínio: Usuario ---
Route::post('/auth/login', 'Domain\\Usuario\\Controller\\AuthController@login',    []);
Route::post('/usuarios',   'Domain\\Usuario\\Controller\\UsuarioController@criar', []);
