DEV_ENV_FILE ?= .env.dev
ENV_FILE ?= $(DEV_ENV_FILE)
WORKTREE_ID ?= $(notdir $(CURDIR))
SUPABASE_DEV_WORKDIR ?= $(HOME)/.cache/go-fiber/supabase/$(WORKTREE_ID)
SUPABASE_DEV_ENV_FILE ?= $(DEV_ENV_FILE)

GOPATH := $(strip $(shell go env GOPATH))
GOBIN := $(strip $(shell go env GOBIN))
ifeq ($(GOBIN),)
GOBIN := $(GOPATH)/bin
endif

AIR_PATH := $(shell command -v air 2>/dev/null)
AIR_BIN := $(if $(AIR_PATH),$(AIR_PATH),$(GOBIN)/air)

LOAD_ENV = set -a; [ -f $(ENV_FILE) ] && . $(ENV_FILE); set +a

.PHONY: dev dev-up dev-stop dev-reset ensure-air ensure-swag regen-docs run build test \
	e2e e2e-base e2e-build e2e-fast e2e-strict e2e-prune e2e-docker-clean \
	e2e-supa e2e-supa-rebuild e2e-supa-base e2e-supa-build e2e-supa-fast e2e-supa-strict e2e-supa-run e2e-supa-prune e2e-supa-docker-clean \
	docker-clean health ci format fly-validate deploy secret-prd

test:
	@go test -race -count=1 $$(go list ./... | grep -v '/test/e2e')

health ci: build test e2e e2e-supa
	@printf "Health check OK\n"

format:
	@go fmt ./...
	@printf "Format OK\n"

dev: dev-up ensure-air regen-docs
	@printf "Running hot-reload server (env: %s)\n" "$(ENV_FILE)"
	@$(LOAD_ENV) && $(AIR_BIN) -c .air.toml

dev-up:
	@./scripts/dev-supabase.sh up --workdir "$(SUPABASE_DEV_WORKDIR)" --env-file "$(SUPABASE_DEV_ENV_FILE)" --worktree-id "$(WORKTREE_ID)"

dev-stop:
	@./scripts/dev-supabase.sh stop --workdir "$(SUPABASE_DEV_WORKDIR)"

dev-reset:
	@./scripts/dev-supabase.sh reset --workdir "$(SUPABASE_DEV_WORKDIR)" --env-file "$(SUPABASE_DEV_ENV_FILE)"

run:
	@$(LOAD_ENV) && go run ./cmd/api

build:
	@mkdir -p tmp
	@go build -o tmp/api ./cmd/api
	@printf "Build OK\n"

ensure-air:
	@if [ ! -x "$(AIR_BIN)" ]; then \
		echo "air not found. Installing..."; \
		GOBIN=$(GOBIN) go install github.com/air-verse/air@latest; \
	fi

ensure-swag regen-docs:
	@$(MAKE) -f make/Makefile.docs.mk $@

fly-validate deploy secret-prd:
	@$(MAKE) -f make/Makefile.fly.mk $@

e2e e2e-base e2e-build e2e-fast e2e-strict e2e-prune:
	@$(MAKE) -f make/Makefile.e2e.mk $@

e2e-docker-clean:
	@$(MAKE) -f make/Makefile.e2e.mk $@

e2e-supa e2e-supa-rebuild e2e-supa-base e2e-supa-build e2e-supa-fast e2e-supa-strict e2e-supa-run e2e-supa-prune e2e-supa-docker-clean:
	@$(MAKE) -f make/Makefile.e2e-supa.mk $@

docker-clean: e2e-docker-clean e2e-supa-docker-clean
	@if command -v supabase >/dev/null 2>&1; then \
		./scripts/dev-supabase.sh reset --workdir "$(SUPABASE_DEV_WORKDIR)" --env-file "$(SUPABASE_DEV_ENV_FILE)"; \
	else \
		printf "supabase CLI not found: skipping dev supabase cleanup\n"; \
	fi
