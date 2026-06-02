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
$response = $curl->post('/auth/login', ['email' => $email, 'senha' => $senha]);
// $response['status'] → HTTP status code
// $response['body']   → array decodificado do JSON da API
```
A URL base é lida de `getenv('API_BASE_URL')` (definida em docker-compose como `http://nginx/api`).

## Fluxo de autenticação
```
login.html
  → auth.js (fetch POST /php/class/Auth/auth.php)
    → auth.php (CurlHelper POST /api/auth/login)
      → API retorna dados do usuário
    → auth.php cria $_SESSION['user']
    → auth.php retorna { status: true }
  → auth.js redireciona para /
```

## Variáveis de ambiente
`API_BASE_URL` é injetada no container front via docker-compose `environment:` e acessada com `getenv('API_BASE_URL')`. O PHP-FPM está configurado com `clear_env = no` para passar variáveis do sistema.

## Convenções
- JS faz `fetch` para arquivos PHP — nunca chama a API diretamente do browser
- Toda resposta PHP retorna `Content-Type: application/json`
- `session_start()` deve ser a primeira instrução de qualquer arquivo PHP que usa sessão
- Não use `echo` fora de arquivos PHP — HTML fica em `.html`
