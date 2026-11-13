# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Regra Principal — Não Violar

**O `cmd/api/main.go` é declarativo e não deve acessar `pkg/` nem `adapter/` diretamente.** Exceção: `pkg/router` é a única chamada permitida de `pkg/` em `main.go` (criação do app Fiber e grupos).

Ao propor mudanças ou refatorações:

1. **Não sugerir** remover camadas de fachada (`cmd/api/mid/`, etc.) para "simplificar" importando `pkg/*` ou `adapter/*` em `main.go`. Essas camadas existem por design.
2. **Não sugerir** adicionar condicionais, regras de negócio ou lógica complexa em `main.go`. O arquivo deve permanecer linear e legível.
3. **Middlewares Fiber e recursos customizáveis** — customizações em `cmd/api/mid/*` (middlewares) ou na fachada correspondente em `cmd/api/*`, não em `main.go` nem diretamente em `pkg/*`.
4. **Novos middlewares ou recursos externos** — criar wrapper em `cmd/api/mid/` (ou nova pasta em `cmd/api/`) delegando para `pkg/*` ou `adapter/*`.
5. **Pacotes de `cmd/api/` não são importados por outros pacotes** — `pkg/`, `adapter/` e `internal/` nunca importam de `cmd/api/*`. Dados fluem via parâmetros (ex: `isProduction bool`).

---

## Comandos

```bash
# Desenvolvimento
make dev              # Hot reload (air + Supabase local)
make run              # go run ./cmd/api
make build            # Compila binário em tmp/api

# Testes
make test             # go test -race -count=1 (exclui test/e2e)
make e2e              # E2E padrão estrito (Postgres + JWKS local)
make e2e-fast         # E2E com cache de build/modules
make e2e-strict       # E2E sem cache
make e2e-supa         # E2E com Supabase local completo (via Docker)
make e2e-supa-fast
make e2e-supa-strict
make health           # build + test + e2e + e2e-supa (pipeline completo)

# Qualidade
make format           # go fmt ./...

# Docs
make regen-docs       # Gera OpenAPI JSON via Swag

# Deploy (Fly.io)
make fly-validate
make secret-prd       # Importa .env.prod para Fly Secrets
make deploy
```

Para rodar um único teste de pacote: `go test -run TestNome ./internal/acl/`

---

## Arquitetura e Fluxo de Inicialização

```
cmd/api/main.go
  → env.Load()                          # Carrega .env
  → bootstrap (logger + DB)             # adapter/log/slog + adapter/db/sqlx
  → pkg/router.Config(...)              # Cria app Fiber
  → cmd/api/mid/* (middlewares)         # cors, helmet, logger, auth, metrics...
  → pkg/router.NewGroup(...) + módulos  # health, hello, acl, docs
  → lifecycle shutdown gracioso         # API, metrics server, DB
```

O DB inicializa de forma permissiva — a API sobe mesmo sem conexão; `/api/readyz` reporta o estado real.

---

## Padrão de Pastas e Responsabilidades

| Camada | Responsabilidade |
|---|---|
| `cmd/api/main.go` | Orquestração declarativa: boot, grupos, lifecycle |
| `cmd/api/mid/*` | Fachadas de middlewares Fiber (auth, cors, metrics, logger…) |
| `cmd/api/*` | Fachadas de recursos externos da API (env, docs, health, acl) |
| `internal/<modulo>` | Domínio por módulo: `model`, `service`, `handler`, `handler_doc`, `route`, `repository`, `queries` |
| `pkg/*` | Componentes genéricos reutilizáveis (auth core, metrics, router, logx, httputil, reqctx…) |
| `adapter/*` | Infraestrutura concreta: `db/sqlx`, `jwt/golangjwt`, `jwks`, `log/slog` |
| `test/e2e/*` | Suíte E2E por features com steps estilo Given/When/Then/And |
| `infra/*` | Dockerfiles e compose para runtime e harness de testes |
| `scripts/*` | Orquestrações shell (dev local, e2e, e2e-supa) |
| `make/*` | Sub-makefiles: `e2e.mk`, `e2e-supa.mk`, `docs.mk`, `fly.mk` |

### Anatomia de um módulo `internal/<modulo>`

```
model.go           # structs de dados
service.go         # lógica de negócio
handler.go         # handlers HTTP (usa service)
handler_doc.go     # anotações Swagger
route.go           # Register(group) – conecta handlers a rotas
repository.go      # contrato da interface de dados
queries.go         # SQL queries
```

Módulos exportam funções de registro chamadas de `main.go`:
```go
hello.PublicRegister(database)
hello.PrivateRegister(database)
```

---

## Semântica de API e Código

- Endpoints públicos e privados em grupos separados (`public` / `private`) montados em `main.go`.
- Erros de auth: envelope JSON `{"error":"<code>"}`.
- No ACL: operações sem permissão retornam `404` (não `403`) — não vaza existência de recurso.
- `/api/livez` e `/api/startupz` validam processo; `/api/readyz` inclui estado do DB.
- Em `APP_ENV=production`, indisponibilidade em `/api/readyz` usa payload mínimo `{"status":"unavailable"}`.
- Implementação genérica → `pkg/*`; customização específica desta API → `cmd/api/*`.
- Evitar estado global fora de infraestrutura explicitamente inicializada.

---

## Padrão de Testes

### Unit e integração

- Ficam próximos ao código (`*_test.go`) em `pkg/*`, `adapter/*`, `internal/*` e `cmd/api/*`.
- Rodar: `make test` (`go test -race -count=1`, excluindo `test/e2e`).

### E2E

Build tag obrigatória: `//go:build e2e`

```
test/e2e/
  e2e_test.go           # TestMain – setup da suíte
  context.go            # TestContext (HTTP client + baseURL)
  features/
    <modulo>/
      public_test.go    # cenários públicos
      auth_test.go      # cenários autenticados
      journey_test.go   # fluxos de usuário complexos
  internal/
    helpers/            # ThenStatusIs, MustParseJSON, AuthHeaders, AndErrorCodeIs
    steps/              # When*, Then*, And* (funções reutilizáveis de ação/asserção)
```

#### Estilo próximo de Gherkin (estrutura obrigatória)

```go
//go:build e2e

// Scenario: public hello endpoint returns 200 with request_id
func TestHelloPublic_Smoke(t *testing.T) {
    t.Run("When_get_public_hello", func(t *testing.T) {
        ctx := e2e.NewTestContext(t)
        resp := steps.WhenGetHelloPublic(t, ctx)

        t.Run("Then_ok", func(t *testing.T) {
            helpers.ThenStatusIs(t, resp, http.StatusOK)
        })
        t.Run("And_request_id_present", func(t *testing.T) {
            parsed := helpers.MustParseJSON(t, resp)
            if rid, _ := parsed["request_id"].(string); rid == "" {
                t.Fatal("request_id should not be empty")
            }
        })
    })
}
```

#### Regras de escrita E2E

1. Fluxo obrigatório: `When → Then → And` via `t.Run(...)` aninhados.
2. Prefixos semânticos nas funções de step: `When*`, `Then*`, `And*`.
3. Priorizar steps reutilizáveis de `internal/steps`; nunca repetir chamadas HTTP cruas nos testes.
4. Asserções comuns ficam em `internal/helpers`.
5. Adicionar comentário `// Scenario: ...` no topo de cada função de teste.
6. Em testes de jornada, documentar pré-requisitos de env (`E2E_OWNER_TOKEN`, etc.) e usar `t.Skip(...)` quando faltarem.

#### Steps: onde criar

| Arquivo | Conteúdo |
|---|---|
| `steps/when.go` | Ações HTTP genéricas (`WhenGetHelloPublic`, `WhenGetHelloMe`…) |
| `steps/then.go` | Asserções sobre resposta (delegam para `helpers`) |
| `steps/and.go` | Asserções/ações encadeadas genéricas |
| `steps/acl_when.go` | Ações específicas do módulo ACL |
| `steps/acl_and.go` | Asserções específicas do módulo ACL |

---

## Infra e Automação

| Arquivo | Propósito |
|---|---|
| `infra/Dockerfile` | Imagem de runtime (multi-stage, distroless) |
| `infra/Dockerfile.test.base` | Base para runners E2E |
| `infra/Dockerfile.test.supa` | Runner E2E com Supabase CLI |
| `infra/docker-compose.test.yml` | Harness padrão E2E (Postgres + runner) |
| `infra/docker-compose.test.strict.yml` | Sem caches de build/modules |
| `infra/docker-compose.test.supa.yml` | Harness isolado para e2e-supa |

Scripts principais em `scripts/`:

- `e2e-run.sh` — orquestra E2E padrão (copy → JWKS/token → API → go test)
- `e2e-supa-run.sh` — orquestra E2E com Supabase completo
- `dev-supabase.sh` — sobe/para/reseta Supabase local e sincroniza `.env.dev`

---

## Checklist Antes de Propor Alterações

- [ ] `main.go` permanece declarativo (sem lógica, sem condicionais)?
- [ ] `main.go` não importa `pkg/*` nem `adapter/*` (exceto `pkg/router`)?
- [ ] Customizações de middleware/recursos externos feitas nas fachadas de `cmd/api/*`?
- [ ] Nenhum pacote fora de `cmd/api/` importa de `cmd/api/*`?
- [ ] Testes adequados atualizados (unit/integration/e2e conforme impacto)?
- [ ] Se afeta fluxo E2E, a organização `features + steps + helpers` foi preservada?
- [ ] Se afeta execução/deploy, `Makefile`, `scripts/` e `infra/` permaneceram coerentes?
