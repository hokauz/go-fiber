WORKTREE_ID ?= $(notdir $(CURDIR))
RUN_ID ?= $(shell date +%Y%m%d-%H%M%S)-$(shell sh -c 'od -An -N2 -tu2 /dev/urandom | tr -d " "')

E2E_BASE_IMAGE ?= go-fiber-test-base
E2E_SUPA_IMAGE ?= go-fiber-supa-test-runner
E2E_SUPA_FORCE_BUILD ?= 0
E2E_BASE_FORCE_BUILD ?= 0
E2E_SUPA_RUN_MODE ?= fast
ifeq ($(origin E2E_SUPA_REUSE), undefined)
ifeq ($(E2E_SUPA_RUN_MODE),strict)
E2E_SUPA_REUSE := 0
else
E2E_SUPA_REUSE := 1
endif
endif
E2E_SUPA_TEST_EMAIL ?= e2e-user@local.test
E2E_SUPA_TEST_PASSWORD ?= Passw0rd!234
E2E_SUPA_VERBOSE ?= 0

.PHONY: e2e-supa e2e-supa-rebuild e2e-supa-base e2e-supa-build e2e-supa-fast e2e-supa-strict e2e-supa-run e2e-supa-prune e2e-supa-docker-clean

# E2E Supabase Docker harness (infra/) - fully isolated from default e2e
e2e-supa: e2e-supa-fast

e2e-supa-rebuild: E2E_SUPA_FORCE_BUILD=1
e2e-supa-rebuild: e2e-supa-fast

e2e-supa-base:
	@if [ "$(E2E_BASE_FORCE_BUILD)" = "1" ] || ! docker image inspect "$(E2E_BASE_IMAGE)" >/dev/null 2>&1; then \
		echo "Building $(E2E_BASE_IMAGE)"; \
		docker build -f infra/Dockerfile.test.base -t "$(E2E_BASE_IMAGE)" .; \
	else \
		echo "Reusing $(E2E_BASE_IMAGE)"; \
	fi

e2e-supa-build: e2e-supa-base
	@if [ "$(E2E_SUPA_FORCE_BUILD)" = "1" ] || ! docker image inspect "$(E2E_SUPA_IMAGE)" >/dev/null 2>&1; then \
		echo "Building $(E2E_SUPA_IMAGE)"; \
		docker build -f infra/Dockerfile.test.supa --build-arg E2E_BASE_IMAGE="$(E2E_BASE_IMAGE)" -t "$(E2E_SUPA_IMAGE)" .; \
	else \
		echo "Reusing $(E2E_SUPA_IMAGE)"; \
	fi

e2e-supa-fast: e2e-supa-build
	@$(MAKE) -f make/Makefile.e2e-supa.mk e2e-supa-run E2E_SUPA_RUN_MODE=fast E2E_SUPA_REUSE=1

e2e-supa-strict: e2e-supa-build
	@$(MAKE) -f make/Makefile.e2e-supa.mk e2e-supa-run E2E_SUPA_RUN_MODE=strict E2E_SUPA_REUSE=0

e2e-supa-run:
	@mkdir -p e2e-report-supa
	@E2E_SUPA_IMAGE="$(E2E_SUPA_IMAGE)" \
		WORKTREE_ID="$(WORKTREE_ID)" \
		RUN_ID="$(RUN_ID)" \
		RUN_MODE="$(E2E_SUPA_RUN_MODE)" \
		SUPABASE_REUSE="$(E2E_SUPA_REUSE)" \
		E2E_TEST_EMAIL="$(E2E_SUPA_TEST_EMAIL)" \
		E2E_TEST_PASSWORD="$(E2E_SUPA_TEST_PASSWORD)" \
		E2E_VERBOSE="$(if $(V),$(V),$(E2E_SUPA_VERBOSE))" \
		docker compose -f infra/docker-compose.test.supa.yml --project-directory . run --rm -T test-runner-supa

e2e-supa-prune:
	@WORKTREE_ID="$(WORKTREE_ID)" \
		E2E_SUPA_IMAGE="$(E2E_SUPA_IMAGE)" \
		docker compose -f infra/docker-compose.test.supa.yml --project-directory . down -v 2>/dev/null || true
	rm -rf e2e-report-supa

e2e-supa-docker-clean:
	@docker image rm -f "$(E2E_SUPA_IMAGE)" >/dev/null 2>&1 || true
	@docker image rm -f "$(E2E_BASE_IMAGE)" >/dev/null 2>&1 || true
	@docker volume rm "go-fiber-supa-mod-$(WORKTREE_ID)" "go-fiber-supa-build-$(WORKTREE_ID)" >/dev/null 2>&1 || true
