# Instruções para Agentes (AI / Automação)

## Regra Principal — Não Violar

**O `cmd/api/main.go` é declarativo e não deve acessar `pkg/` nem `adapter/` diretamente.** Exceção: `pkg/router` é a única chamada permitida de `pkg/` em `main.go` (criação do app Fiber e grupos).

Ao propor mudanças ou refatorações:

1. **Não sugerir** remover camadas de fachada (`cmd/api/mid/`, etc.) para "simplificar" importando `pkg/*` ou `adapter/*` em `main.go`. Essas camadas existem por design.

2. **Não sugerir** adicionar condicionais, regras de negócio ou lógica complexa em `main.go`. O arquivo deve permanecer linear e legível.

3. **Middlewares Fiber e recursos customizáveis** — Customizações devem ser feitas em `cmd/api/mid/*` (middlewares) ou na fachada correspondente em `cmd/api/*` (router, docs, env, etc.), não em `main.go` nem diretamente em `pkg/*` quando a mudança for específica desta API.

4. **Novos middlewares ou recursos externos** — Criar o wrapper em `cmd/api/mid/` (ou nova pasta em `cmd/api/` se aplicável) que delega ou customiza a partir de `pkg/*` ou `adapter/*`, em vez de importar diretamente no `main.go`.

5. **Pacotes de `cmd/api/` não devem ser importados por outros pacotes** — Pacotes em `pkg/`, `adapter/` e `internal/` nunca importam de `cmd/api/*`. Dados que precisam fluir de `cmd/api/` para camadas inferiores devem ser passados como parâmetros (ex: `isProduction bool`), não por importação direta.

## Padrão de Pastas e Responsabilidades

| Camada/Pasta | Responsabilidade |
|---|---|
| `cmd/api/main.go` | Orquestração declarativa da API: ordem de boot, montagem de grupos e startup lifecycle |
| `cmd/api/mid/*` | Fachadas de middlewares Fiber (auth, cors, metrics, logger, etc.) |
| `cmd/api/*` | Fachadas de recursos externos da API (`env`, `docs`, `health`, `acl`) |
| `internal/<modulo>` | Domínio da aplicação por módulo (model, service, handler, route, queries, repository contracts) |
| `pkg/*` | Componentes genéricos reutilizáveis (auth core, metrics, envloader, router, log abstractions, utils) |
| `adapter/*` | Infraestrutura concreta para ports (`db/sqlx`, `jwt`, `jwks`, `log/slog`) |
| `test/e2e/*` | Suíte E2E por features, com steps reutilizáveis estilo Given/When/Then/And |
| `infra/*` | Dockerfiles e compose para runtime e harness de testes |
| `scripts/*` | Orquestrações shell para dev local, e2e e e2e-supa |
| `make/*` | Sub-makefiles especializados (`e2e`, `e2e-supa`, `docs`, `fly`) |

## Semântica e Convenções do Projeto

### Semântica de API

- Endpoints públicos e privados são registrados em grupos separados (`public` e `private`) no `main.go`.
- Falhas de autenticação retornam envelope JSON no formato `{"error":"<code>"}`.
- No ACL, operações sem permissão retornam `404` (não `403`) para não vazar existência de recurso.
- `/api/livez` e `/api/startupz` validam processo; `/api/readyz` representa prontidão (inclui estado de DB).
- Em `APP_ENV=production`, indisponibilidade em `/api/readyz` usa payload mínimo (`{"status":"unavailable"}`).

### Semântica de Código

- `main.go` deve permanecer linear: carregar env, criar logger/db, configurar app, registrar grupos, iniciar lifecycle.
- Implementação genérica vai em `pkg/*`; customização específica desta API vai em `cmd/api/*`.
- Adapters devem implementar portas dos pacotes de domínio/genéricos, sem acoplamento inverso.
- Evitar estado global fora de infraestrutura explicitamente controlada (logger default é bootstrap explícito).

## Arquitetura (Fluxo de Inicialização)

1. Carrega env com `cmd/api/env.Load()`.
2. Inicializa logger base (`adapter/log/slog`) e registra no `pkg/logx`.
3. Inicializa DB (`adapter/db/sqlx`) de forma permissiva; app continua sem DB quando indisponível.
4. Cria app Fiber via `pkg/router.Config(...)`.
5. Registra middlewares por fachadas `cmd/api/mid/*`.
6. Cria grupos de rota com `pkg/router.NewGroup(...)` e registra módulos (`health`, `hello`, `acl`, `docs`).
7. Executa lifecycle com shutdown gracioso de API, metrics server e DB.

## Padrão de Testes

### Unit e integração leve

- Testes ficam próximos ao código (`*_test.go`) em `pkg/*`, `adapter/*`, `internal/*` e `cmd/api/*`.
- Comando padrão: `make test`.
- `make test` executa `go test -race -count=1` excluindo `test/e2e`.

### E2E

- Todos os E2E usam build tag `e2e`.
- Entrypoint de suíte em `test/e2e/e2e_test.go`.
- Contexto compartilhado HTTP em `test/e2e/context.go`.
- Features por módulo em `test/e2e/features/<modulo>`.
- Helpers e steps compartilhados em `test/e2e/internal/helpers` e `test/e2e/internal/steps`.

Comandos:

- `make e2e` (default: strict com Postgres + JWKS local)
- `make e2e-fast`
- `make e2e-strict`
- `make e2e-supa` (Supabase local via Docker)
- `make e2e-supa-fast`
- `make e2e-supa-strict`

## Organização E2E (estilo próximo de Gherkin)

Estrutura esperada:

```text
test/e2e/
  features/
    hello/
    acl/
  internal/
    helpers/
    steps/
```

Regras de escrita:

1. Cenários em forma de fluxo `When -> Then -> And` com `t.Run(...)`.
2. Funções de steps devem usar prefixos semânticos (`When...`, `Then...`, `And...`).
3. Features devem priorizar steps reutilizáveis em vez de repetir chamadas HTTP cruas.
4. Asserções comuns ficam em `internal/helpers`.
5. Em testes de jornada, documentar pré-requisitos de env e usar `t.Skip(...)` quando faltarem variáveis.
6. Manter comentários `Scenario:` no topo dos testes de feature para leitura humana.

## Infra e Automação

- `infra/Dockerfile`: imagem de runtime da API.
- `infra/Dockerfile.test.base`: base comum para runners E2E.
- `infra/Dockerfile.test.supa`: runner E2E com Supabase CLI.
- `infra/docker-compose.test.yml`: harness padrão E2E.
- `infra/docker-compose.test.strict.yml`: modo estrito sem caches de build/modules.
- `infra/docker-compose.test.supa.yml`: harness isolado para e2e-supa.

Scripts principais:

- `scripts/dev-supabase.sh`: sobe/para/reseta Supabase local e sincroniza `.env.dev`.
- `scripts/e2e-run.sh`: orquestra e2e padrão (copy -> token/JWKS -> API -> go test).
- `scripts/e2e-supa-run.sh`: orquestra e2e com Supabase local completo.

## Checklist Antes de Propor Alterações

- [ ] A alteração mantém `main.go` declarativo?
- [ ] `main.go` continua sem importar `pkg/*` nem `adapter/*` (exceto `pkg/router`)?
- [ ] Customizações de middleware/recursos externos foram feitas nas fachadas de `cmd/api/*`?
- [ ] Nenhum pacote fora de `cmd/api/` importa de `cmd/api/*`?
- [ ] Testes adequados foram atualizados (unit/integration/e2e conforme impacto)?
- [ ] Se afetar fluxo E2E, a organização `features + steps + helpers` foi preservada?
- [ ] Se afetar execução/deploy, `Makefile`, `scripts/` e `infra/` permaneceram coerentes?
