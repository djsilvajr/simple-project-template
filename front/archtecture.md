# Arquitetura do Frontend

## Visão geral das camadas

```
Browser
  │
  ├── GET /index.html, /css/*, /js/*   ← nginx serve estático de /var/www/front
  │
  └── GET/POST /*.php                  ← nginx → front container PHP-FPM :9000
        │
        ├── session_check.php          lê $_SESSION, retorna JSON
        ├── php/class/Auth/auth.php    recebe credenciais, chama API, salva sessão
        ├── php/class/Auth/logout.php  destrói sessão
        └── php/class/helper/CurlHelper.php
              │
              └── HTTP POST http://nginx/api/*   ← nginx roteia para back container
```

## Fluxo de uma página protegida

```
1. Browser carrega auth-needed.html
2. session-check.js executa (IIFE async)
   a. Oculta <html> (visibility: hidden)
   b. fetch GET /php/class/session_check.php
   c. session_check.php lê $_SESSION['user']
      → não autenticado: 302 para /Auth/login.html
      → autenticado:     expõe window.__session, mostra página
```

## Fluxo de autenticação

```
login.html
  └── auth.js
        └── fetch POST /php/class/Auth/auth.php  (JSON: {email, senha})
              └── auth.php
                    └── CurlHelper.post('/auth/login', {email, senha})
                          └── curl POST http://nginx/api/auth/login
                                └── nginx → back PHP-FPM
                                      └── AuthController → UseCase → Repository → MySQL
                                            ← {status:true, data:{id,usuario,email,...}}
                    └── $_SESSION['user'] = $response['body']['data']
                    ← {status: true}
              ← {status: true}
        └── window.location = '/'
```

## Estrutura de pastas

```
front/
├── index.html                 página inicial pública
├── Auth/
│   └── login.html             formulário de login
├── test/
│   ├── auth-needed.html       exemplo de página protegida
│   └── non-auth-needed.html   exemplo de página pública
├── css/
│   ├── main.css               reset global + estilos de páginas públicas
│   └── Auth/
│       └── auth.css           estilos específicos da tela de login
├── js/
│   ├── main.js                utilitários globais (logout, loadUserInfo)
│   └── Auth/
│       ├── auth.js            lógica do formulário de login
│       └── session-check.js   guarda de rota (IIFE — redireciona se não autenticado)
└── php/
    └── class/
        ├── Auth/
        │   ├── auth.php       endpoint de login (cria sessão)
        │   └── logout.php     endpoint de logout (destrói sessão)
        ├── session_check.php  endpoint de verificação de sessão
        └── helper/
            └── CurlHelper.php wrapper de cURL para a API
```

## Regras de organização

| O que criar | Onde colocar |
|---|---|
| Nova página protegida | `NomeFeature/pagina.html` + incluir `session-check.js` |
| Nova página pública | `NomeFeature/pagina.html` sem `session-check.js` |
| JS específico da feature | `js/NomeFeature/feature.js` |
| CSS específico da feature | `css/NomeFeature/feature.css` |
| PHP que chama API | `php/class/NomeFeature/endpoint.php` |

## Sessão

A sessão é gerenciada inteiramente pelo PHP do container front. O browser mantém o cookie `PHPSESSID` e o envia automaticamente em todas as requisições para arquivos `.php`. O JS nunca manipula o cookie diretamente.
