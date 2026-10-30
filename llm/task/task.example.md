# Como usar IA para desenvolvimento neste projeto

Este arquivo mostra como estruturar tarefas para obter o melhor resultado da IA ao desenvolver no backend ou frontend deste template.

---

## Princípios gerais

1. **Sempre forneça o contexto relevante** — indique qual arquivo de contexto a IA deve ler antes de começar (`back/context.md`, `front/context.md`, `llm/context/context.md`)
2. **Referencie o exemplo existente** — o domínio `back/internal/domain/user` no backend e as páginas `test/` no frontend são referências funcionais
3. **Delimite o escopo** — diga exatamente o que deve ser criado ou modificado, sem deixar margem para invenção
4. **Especifique o padrão esperado** — o projeto tem padrões definidos; mencione-os explicitamente para evitar desvios

---

## Template de task — Backend

```
Leia `back/context.md` e `back/archtecture.md` antes de começar.
Siga exatamente o padrão do domínio `back/internal/domain/user` como referência.

## Objetivo
[Descrição clara do que precisa ser feito]

## Rota(s)
- METHOD /caminho → handler.Metodo [middlewares: rate_limit | auth | admin]

## Model envolvido
[Struct GORM existente, ou "criar novo model" com o struct completo]

## Regras de negócio
- [Validação 1]
- [Validação 2]
- [Regra de conflito, se houver]

## Resposta esperada (sucesso)
```json
{
  "campo": "valor"
}
```

## Arquivos a criar/modificar
- back/internal/domain/<nome>/model/<nome>.go      ← entidade + DTOs
- back/internal/domain/<nome>/errors/errors.go
- back/internal/domain/<nome>/repository/repository.go
- back/internal/domain/<nome>/service/service.go
- back/internal/domain/<nome>/usecase/<caso_de_uso>.go
- back/internal/domain/<nome>/handler/handler.go
- back/internal/router/router.go   ← instanciar dependências e registrar rotas
- back/migration/migrate.go        ← adicionar o model em `models` (se for novo)
```

---

## Template de task — Frontend

```
Leia `front/context.md` e `front/archtecture.md` antes de começar.
Siga o padrão das páginas existentes: `Auth/login.html` (protegida) e `test/non-auth-needed.html` (pública).

## Objetivo
[Descrição clara do que precisa ser feito]

## Página
- Caminho: /NomeFeature/pagina.html
- Requer autenticação: sim | não

## Comportamento esperado
- [O que o usuário vê/faz]
- [Chamadas PHP que o JS faz]
- [Redirecionamentos, se houver]

## Endpoint PHP (se necessário)
- Arquivo: php/class/NomeFeature/endpoint.php
- Chama: API /api/rota via CurlHelper

## Arquivos a criar/modificar
- front/NomeFeature/pagina.html
- front/js/NomeFeature/feature.js
- front/css/NomeFeature/feature.css
- front/php/class/NomeFeature/endpoint.php  (se necessário)
```

---

## Exemplos reais

### Exemplo 1 — Backend: buscar usuário por login

```
Leia `back/context.md`. Siga o padrão de `back/internal/domain/user`.

## Objetivo
Criar rota para um admin buscar um usuário pelo login.

## Rota
- GET /users/by-login/{login} → handler.GetByLogin [middlewares: rate_limit, auth, admin]

## Model
user.User (já existe) — o campo Password nunca é serializado (json:"-")

## Regras de negócio
- Login inexistente → 404 (usererrors.ErrNotFound)
- Apenas admin (use case retorna ErrForbidden caso contrário)

## Resposta esperada
{ "status": true, "statusCode": 200, "data": { "id": 1, "login": "...", ... } }

## Arquivos a criar/modificar
- back/internal/domain/user/service/service.go      ← adicionar GetByLogin()
- back/internal/domain/user/usecase/get_user_by_login.go
- back/internal/domain/user/handler/handler.go      ← adicionar método GetByLogin
- back/internal/router/router.go                    ← registrar a rota
```

---

### Exemplo 2 — Frontend: página de perfil protegida

```
Leia `front/context.md`. Siga o padrão de `test/auth-needed.html`.

## Objetivo
Criar página de perfil que exibe os dados do usuário logado.

## Página
- Caminho: /perfil/index.html
- Requer autenticação: sim (incluir session-check.js)

## Comportamento
- Ao carregar, session-check.js verifica sessão
- Os dados de window.__session são exibidos na página (login, name, email)
- Botão de logout chama logout() de main.js

## Arquivos a criar
- front/perfil/index.html
- front/js/perfil/perfil.js   ← exibe window.__session na página
- front/css/perfil/perfil.css ← estilos do card de perfil
```

---

## Dicas de eficiência

- **Não peça "crie tudo para um CRUD"** — divida em tarefas menores (criar, listar, atualizar, deletar) para manter controle
- **Mencione erros esperados** — diga qual HTTP code e mensagem cada cenário de erro deve retornar
- **Referencie arquivos existentes** — "siga o padrão de X" é mais eficiente do que descrever o padrão do zero
- **Uma task por domínio** — não misture backend e frontend na mesma task; são contextos diferentes
- **Após a task**, revise os arquivos gerados, confirme que as rotas foram registradas em `router.go`, que models novos estão em `migration/migrate.go` e rode `go vet ./... && go build ./...` em `back/`
