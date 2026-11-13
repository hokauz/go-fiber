# Setup Linux

Este guia prepara o ambiente para rodar `dev`, `test`, `build`, `e2e` e `e2e-supa`.

## 1. Instalar dependências

### 1.1 Debian/Ubuntu (exemplo)

```bash
sudo apt update
sudo apt install -y make curl rsync jq ca-certificates gnupg lsb-release
```

Go (ajuste para a versão do `go.mod`):

```bash
sudo apt install -y golang-go
```

Se a versão do pacote da distro estiver abaixo da exigida no `go.mod`, instale Go manualmente pelo tarball oficial e confirme com:

```bash
go version
```

### 1.2 Docker + Compose plugin

Siga o fluxo oficial da sua distro e valide:

```bash
docker version
docker compose version
```

No Linux, normalmente é necessário adicionar o usuário ao grupo `docker`:

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### 1.3 Supabase CLI (binário)

Exemplo para Linux `amd64`:

```bash
curl -fsSL https://github.com/supabase/cli/releases/download/v2.75.0/supabase_linux_amd64.tar.gz -o /tmp/supabase.tgz
tar -xzf /tmp/supabase.tgz -C /tmp
sudo mv /tmp/supabase /usr/local/bin/supabase
sudo chmod +x /usr/local/bin/supabase
supabase --version
```

Opcional (deploy Fly):

- instale `flyctl` pelo método oficial da sua distro.

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

Rebuild forçado de imagens de teste:

```bash
make e2e E2E_FORCE_BUILD=1
make e2e-supa E2E_SUPA_FORCE_BUILD=1
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

## 6. Problemas comuns no Linux

### `permission denied` no Docker

- confirme que o serviço Docker está ativo
- confirme usuário no grupo `docker`

### E2E Supabase falhando por networking

- valide se o daemon Docker e o `docker compose` estão saudáveis
- tente `make e2e-supa E2E_SUPA_FORCE_BUILD=1`
