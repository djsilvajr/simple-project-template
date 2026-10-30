# Contexto do Frontend

## Stack
- HTML5, CSS3, JavaScript ES2022 — sem framework, sem bundler
- PHP 8.2 usado exclusivamente para: sessão (`$_SESSION`) e chamadas à API via cURL
- Sem templating PHP — HTML é puro, PHP é chamado via fetch/XMLHttpRequest

## Organização de arquivos
CSS e JS seguem a mesma estrutura de diretórios:

```
css/                    js/
├── main.css            ├── main.js          global / páginas públicas
└── Auth/               └── Auth/
    └── auth.css            ├── auth.js      lógica do formulário de login
                            └── session-check.js  guarda de rota
```

Crie sempre o CSS e o JS da feature no mesmo subdiretório (`Feature/feature.css` e `Feature/feature.js`).

## Controle de acesso (sessão)

### Página protegida
Inclua `session-check.js` como **primeiro** `<script>` no `<head>`:
```html
<head>
    ...
    <script src="/js/Auth/session-check.js"></script>
</head>
```
O script oculta a página imediatamente, verifica a sessão via `session_check.php` e redireciona para `/Auth/login.html` se não autenticado. Após confirmação expõe `window.__session` com os dados do usuário.

### Página pública
Não inclua `session-check.js`. Use `main.js` normalmente.

## PHP — responsabilidades
| Arquivo | Função |
|---|---|
| `php/class/Auth/auth.php` | Recebe credenciais do JS, chama API via cURL, cria `$_SESSION['user']` |
| `php/class/Auth/logout.php` | Destrói a sessão |
| `php/class/session_check.php` | Retorna `{ autenticado, usuario }` para o JS |
| `php/class/helper/CurlHelper.php` | Wrapper de cURL para chamar a API |

## CurlHelper — uso
```php
require_once __DIR__ . '/../helper/CurlHelper.php';

$curl     = new CurlHelper();
$response = $curl->post('/auth/login', ['login' => $login, 'password' => $password]);
// $response['status'] → HTTP status code
// $response['body']   → array decodificado do JSON da API

// Rotas protegidas: passe o JWT salvo na sessão
$curl = new CurlHelper($_SESSION['token']);
$me   = $curl->get('/auth/me');
```
A URL base é lida de `getenv('API_BASE_URL')` e o timeout de `getenv('API_TIMEOUT')` (ambos em `front/.env`).

## Fluxo de autenticação
```
login.html
  → auth.js (fetch POST /php/class/Auth/auth.php)
    → auth.php (CurlHelper POST /api/auth/login com {login, password})
      → API retorna { token, expiresAt, user }
    → auth.php cria $_SESSION['user'] e $_SESSION['token']
    → auth.php retorna { status: true }
  → auth.js redireciona para /
```

## Variáveis de ambiente
Definidas em `front/.env` (copie de `front/.env.example`), carregadas no container via `env_file:` do docker-compose e lidas com `getenv()`:

| Variável | Uso |
|---|---|
| `API_BASE_URL` | URL base da API (`http://nginx/api` dentro da rede Docker) |
| `API_TIMEOUT` | Timeout, em segundos, das chamadas cURL à API |
| `SESSION_LIFETIME` | Vida da sessão em minutos (igual ao `JWT_EXPIRATION_MINUTES` da API) |
| `SESSION_REFRESH_GRACE` | Minutos após a expiração do token em que o front tenta renová-lo (igual ao `JWT_REFRESH_GRACE_MINUTES` da API) |
| `SESSION_SECURE_COOKIE` | Força o flag `Secure` no cookie (ligado automaticamente sob HTTPS) |
 O PHP-FPM está configurado com `clear_env = no` para passar variáveis do sistema.

## Convenções
- JS faz `fetch` para arquivos PHP — nunca chama a API diretamente do browser
- Toda resposta PHP retorna `Content-Type: application/json`
- Nunca chame `session_start()` direto: use `Session::start()` (`php/class/helper/Session.php`) como primeira instrução de qualquer arquivo PHP que usa sessão. Ele aplica as flags do cookie (`HttpOnly`, `SameSite=Lax`, `Secure`) e o tempo de vida. Se o token da API expirou há no máximo `SESSION_REFRESH_GRACE` minutos, renova em `POST /auth/refresh` e salva o novo token; senão, descarta a sessão
- Logout: `Session::destroy()` limpa os dados, o arquivo da sessão e o cookie do navegador
- Não use `echo` fora de arquivos PHP — HTML fica em `.html`
