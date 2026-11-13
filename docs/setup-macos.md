# Setup macOS

Este guia prepara o ambiente para rodar `dev`, `test`, `build`, `e2e` e `e2e-supa`.

## 1. Instalar dependências

### 1.1 Homebrew (se ainda não tiver)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### 1.2 Ferramentas base

```bash
brew install go make curl
```

### 1.3 Docker Desktop

- Instale Docker Desktop para macOS.
- Verifique:

```bash
docker version
docker compose version
```

### 1.4 Supabase CLI

```bash
brew install supabase/tap/supabase
supabase --version
```

Opcional (deploy):

```bash
brew install flyctl
```

## 2. Preparar projeto

```bash
cp .env.example .env.dev
```

## 3. Executar em desenvolvimento

```bash
make dev
```

Em outro terminal:

```bash
curl -i http://127.0.0.1:8080/api/livez
curl -i http://127.0.0.1:8080/api/readyz
```

Parar/resetar ambiente:

```bash
make dev-stop
make dev-reset
```

## 4. Rodar testes

Unit/integration:

```bash
make test
```

E2E padrão:

```bash
make e2e
```

E2E Supabase:

```bash
make e2e-supa
```

Verbose:

```bash
make e2e V=1
make e2e-supa V=1
```

## 5. Build e run

```bash
make build
./tmp/api
```

Ou:

```bash
make run
```

## 6. Problemas comuns no macOS

### Docker sem acesso

Se aparecer erro de permissão no socket Docker:

- confirme Docker Desktop aberto
- reinicie Docker Desktop e tente novamente

### Cache de runner E2E desatualizado

```bash
make e2e E2E_FORCE_BUILD=1
make e2e-supa E2E_SUPA_FORCE_BUILD=1
```
