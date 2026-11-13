# Go Fiber Template

Template de API em Go + Fiber com arquitetura em camadas, middlewares desacoplados, health checks, métricas, ACL e suíte E2E com dois modos (`e2e` e `e2e-supa`).

## 1. Objetivo deste README

Este arquivo cobre o tutorial completo para executar o projeto:

- desenvolvimento local (`make dev`, `make run`)
- testes (`make test`, `make e2e`, `make e2e-supa`)
- build (`make build`)
- geração de docs (`make regen-docs`)
- deploy com Fly (`make deploy`)

## 2. Setup por sistema operacional

Para evitar um README gigante, o setup foi separado por ambiente:

- [Setup macOS](docs/setup-macos.md)
- [Setup Linux](docs/setup-linux.md)
- [Setup Windows (WSL2 recomendado)](docs/setup-windows.md)

## 3. Requisitos comuns

- Go (versão do `go.mod`, atualmente `1.26.x`)
- `make`
- Docker + Docker Compose
- `curl`
- Supabase CLI (`supabase`) para fluxo `make dev`

Opcional:

- `flyctl` para deploy

## 4. Preparar o projeto

No diretório raiz:

```bash
cp .env.example .env.dev
```

Observações:

- O `make dev` usa por padrão `ENV_FILE=.env.dev`.
- O script `scripts/dev-supabase.sh up` atualiza `.env.dev` com `DB_URL`, `AUTH_JWKS_URL` e variáveis `SUPABASE_*` de desenvolvimento.

## 5. Fluxo rápido (primeira execução)

```bash
make dev
```

Esse comando:

1. sobe Supabase local (`make dev-up`)
2. instala `air` se necessário
3. regenera docs OpenAPI
4. inicia API com hot reload

Health checks:

```bash
curl -i http://127.0.0.1:8080/api/livez
curl -i http://127.0.0.1:8080/api/startupz
curl -i http://127.0.0.1:8080/api/readyz
```

Parar/resetar ambiente dev:

```bash
make dev-stop
make dev-reset
```

## 6. Comandos principais

| Comando | Uso |
|---|---|
| `make dev` | Sobe Supabase local + hot reload da API |
| `make dev-up` | Sobe apenas Supabase local e sincroniza `.env.dev` |
| `make dev-stop` | Para stack Supabase local |
| `make dev-reset` | Remove stack Supabase local e limpa env gerenciado |
| `make run` | Executa `go run ./cmd/api` com env atual |
| `make build` | Gera binário em `tmp/api` |
| `make test` | Testes unit/integration com `-race`, sem E2E |
| `make e2e` | E2E padrão (modo strict) com Postgres + JWKS local |
| `make e2e-fast` | E2E padrão com volumes de cache |
| `make e2e-strict` | E2E padrão sem cache (mais isolado) |
| `make e2e-supa` | E2E com Supabase local (modo fast) |
| `make e2e-supa-fast` | E2E Supabase com reuso de stack |
| `make e2e-supa-strict` | E2E Supabase com stack isolada por execução |
| `make regen-docs` | Regenera `docs/openapi.json` |
| `make fly-validate` | Valida `fly.toml` |
| `make secret-prd` | Importa `.env.prod` para Fly Secrets |
| `make deploy` | Deploy no Fly |
| `make docker-clean` | Limpa artefatos docker de e2e/e2e-supa + reset supabase dev |

## 7. Testes

### 7.1 Testes unit/integration

```bash
make test
```

Equivalente ao padrão do projeto:

```bash
go test -race -count=1 $(go list ./... | grep -v '/test/e2e')
```

### 7.2 E2E padrão (`make e2e`)

```bash
make e2e
```

Características:

- sobe Postgres via Docker Compose (`infra/docker-compose.test.yml`)
- gera JWKS + token localmente no runner
- aplica migrations
- sobe API e roda `go test -tags=e2e ./test/e2e/...`

Rebuild forçado de imagem runner:

```bash
make e2e E2E_FORCE_BUILD=1
```

### 7.3 E2E com Supabase (`make e2e-supa`)

```bash
make e2e-supa
```

Características:

- runner com Supabase CLI
- sobe stack Supabase local
- cria usuário de teste e token real de auth
- aplica migrations
- sobe API e executa suíte `-tags=e2e`

Rebuild forçado:

```bash
make e2e-supa E2E_SUPA_FORCE_BUILD=1
```

Variáveis úteis:

- `E2E_SUPA_TEST_EMAIL` (default `e2e-user@local.test`)
- `E2E_SUPA_TEST_PASSWORD` (default `Passw0rd!234`)
- `V=1` para verbose

### 7.4 Testes de jornada ACL (opcionais)

Alguns cenários em `test/e2e/features/acl/journey_test.go` exigem variáveis:

- `E2E_OWNER_TOKEN`
- `E2E_MEMBER_TOKEN`
- `E2E_RESOURCE_ID`
- `E2E_RESOURCE_TYPE` (opcional, default `account`)

Se não estiverem definidas, os testes fazem `t.Skip`.

## 8. Build e execução sem hot reload

```bash
make build
./tmp/api
```

Ou diretamente:

```bash
make run
```

## 9. Documentação OpenAPI e rota `/api/docs`

Regenerar docs:

```bash
make regen-docs
```

Arquivos gerados:

- `docs/swagger.json` (intermediário)
- `docs/openapi.json` (final)

A rota `/api/docs`:

- fica habilitada fora de produção
- é desabilitada quando `APP_ENV=production`

## 10. Estrutura de pastas (resumo)

```text
cmd/api/        # bootstrap declarativo e fachadas da API
internal/       # domínio por módulo (hello, acl, ...)
pkg/            # componentes genéricos reutilizáveis
adapter/        # infraestrutura concreta (db, jwt, jwks, log)
infra/          # Dockerfiles e compose
test/e2e/       # suíte E2E por features + steps/helpers
scripts/        # automações shell (dev/e2e/e2e-supa)
make/           # sub-makefiles (e2e/e2e-supa/docs/fly)
```

Regras arquiteturais para agentes e colaboradores: [AGENTS.md](AGENTS.md).

## 11. Deploy Fly.io

### 11.1 Preparar `.env.prod`

Baseie em `.env.example` e ajuste no mínimo:

- `APP_ENV=production`
- `CORS_ORIGINS=...`
- `DB_URL=...`
- `AUTH_ISSUER`, `AUTH_AUDIENCE`, `AUTH_JWKS_URL` (se rotas privadas)

### 11.2 Comandos

```bash
make secret-prd
make fly-validate
make deploy
```

### 11.3 Smoke pós deploy

```bash
curl -i https://<seu-app>.fly.dev/api/livez
curl -i https://<seu-app>.fly.dev/api/readyz
fly status -a <seu-app>
fly logs -a <seu-app>
```

## 12. Troubleshooting rápido

### 12.1 Docker socket sem permissão

Erro típico:

```text
permission denied while trying to connect to the docker API
```

Ação:

- garanta que Docker está ativo
- no Linux/WSL, valide permissão do usuário no daemon Docker

### 12.2 Runner E2E desatualizado por cache

Quando scripts mudam e o runner reaproveita imagem antiga:

```bash
make e2e E2E_FORCE_BUILD=1
make e2e-supa E2E_SUPA_FORCE_BUILD=1
```

### 12.3 `/api/readyz` retornando `503`

Causa comum:

- `DB_URL` ausente ou DB indisponível

Comportamento esperado:

- em produção: payload mínimo `{"status":"unavailable"}`
- em outros ambientes: payload detalhado com `reason`
