# Setup Windows

Para este projeto, o caminho recomendado é **WSL2 + Ubuntu**.

Motivo: os scripts (`scripts/*.sh`) e os alvos `make` foram desenhados para ambiente Unix.

## 1. Pré-requisitos no Windows

- WSL2 habilitado
- Ubuntu instalado no WSL
- Docker Desktop com integração WSL ativada
- Git

## 2. Configurar Ubuntu (WSL)

Abra o terminal Ubuntu e instale as dependências:

```bash
sudo apt update
sudo apt install -y make curl rsync jq ca-certificates
sudo apt install -y golang-go
```

Se o `golang-go` do Ubuntu estiver abaixo da versão exigida no `go.mod`, instale Go manualmente no WSL e valide:

```bash
go version
```

Valide Docker via integração WSL:

```bash
docker version
docker compose version
```

## 3. Instalar Supabase CLI no WSL

Exemplo para Linux `amd64`:

```bash
curl -fsSL https://github.com/supabase/cli/releases/download/v2.75.0/supabase_linux_amd64.tar.gz -o /tmp/supabase.tgz
tar -xzf /tmp/supabase.tgz -C /tmp
sudo mv /tmp/supabase /usr/local/bin/supabase
sudo chmod +x /usr/local/bin/supabase
supabase --version
```

## 4. Abrir o projeto no WSL

Use um caminho Linux (evite `/mnt/c/...` para melhor performance de I/O), por exemplo:

```bash
cd ~/workspace/templates/go-fiber
cp .env.example .env.dev
```

## 5. Executar em desenvolvimento

```bash
make dev
```

Em outro terminal WSL:

```bash
curl -i http://127.0.0.1:8080/api/livez
curl -i http://127.0.0.1:8080/api/readyz
```

Parar/resetar:

```bash
make dev-stop
make dev-reset
```

## 6. Rodar testes e build

```bash
make test
make e2e
make e2e-supa
make build
./tmp/api
```

## 7. Problemas comuns no Windows/WSL

### Docker indisponível no WSL

- abra Docker Desktop
- ative integração com a distro Ubuntu em `Settings > Resources > WSL Integration`

### Scripts lentos ou comportamento estranho em `/mnt/c`

- mova o projeto para filesystem Linux da distro (`/home/<usuario>/...`)

### Cache de runner desatualizado

```bash
make e2e E2E_FORCE_BUILD=1
make e2e-supa E2E_SUPA_FORCE_BUILD=1
```
