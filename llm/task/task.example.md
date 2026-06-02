# Como usar IA para desenvolvimento neste projeto

Este arquivo mostra como estruturar tarefas para obter o melhor resultado da IA ao desenvolver no backend ou frontend deste template.

---

## Princípios gerais

1. **Sempre forneça o contexto relevante** — indique qual arquivo de contexto a IA deve ler antes de começar (`back/context.md`, `front/context.md`, `llm/context/context.md`)
2. **Referencie o exemplo existente** — o domínio `Domain/Teste` no backend e as páginas `test/` no frontend são referências funcionais
3. **Delimite o escopo** — diga exatamente o que deve ser criado ou modificado, sem deixar margem para invenção
4. **Especifique o padrão esperado** — o projeto tem padrões definidos; mencione-os explicitamente para evitar desvios

---

## Template de task — Backend

```
Leia `back/context.md` e `back/archtecture.md` antes de começar.
Siga exatamente o padrão do domínio `Domain/Teste` como referência.

## Objetivo
[Descrição clara do que precisa ser feito]

## Rota(s)
- METHOD /caminho → Controller@metodo [middleware: none|auth]

## Tabela envolvida
[Nome da tabela e colunas relevantes, ou "criar nova tabela" com a estrutura]

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
- back/src/Domain/NomeDominio/DTO/NomeDTO.php
- back/src/Domain/NomeDominio/Repository/NomeRepositoryInterface.php
- back/src/Domain/NomeDominio/Repository/NomeRepository.php
- back/src/Domain/NomeDominio/UseCase/NomeUseCase.php
- back/src/Domain/NomeDominio/Controller/NomeController.php
- back/src/Routes/main.php  ← adicionar rota
- db/create_nome_tabela.sql  ← se criar tabela nova
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

### Exemplo 1 — Backend: listar usuários

```
Leia `back/context.md`. Siga o padrão de `Domain/Teste`.

## Objetivo
Criar rota para listar todos os usuários cadastrados.

## Rota
- GET /usuarios → Domain\Teste\Controller\UsuarioController@listar [sem middleware]

## Tabela
usuarios (id, usuario, email, criacao_datahora, alteracao_datahora) — sem retornar senha

## Regras de negócio
- Nenhuma validação de entrada (GET sem parâmetros)
- Retornar array de usuários sem o campo senha

## Resposta esperada
{ "status": true, "statusCode": 200, "data": [ { "id": 1, "usuario": "..." } ] }

## Arquivos a modificar
- back/src/Domain/Teste/Repository/UsuarioRepositoryInterface.php  ← adicionar listar()
- back/src/Domain/Teste/Repository/UsuarioRepository.php          ← implementar listar()
- back/src/Domain/Teste/Controller/UsuarioController.php          ← adicionar método listar()
- back/src/Routes/main.php                                         ← adicionar Route::get
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
- Os dados de window.__session são exibidos na página (usuario, email)
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
- **Após a task**, revise os arquivos gerados e confirme que as rotas foram adicionadas em `main.php` e que o SQL foi criado em `db/`
