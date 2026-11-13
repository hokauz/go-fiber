WORKTREE_ID ?= $(notdir $(CURDIR))
E2E_BASE_IMAGE ?= go-fiber-test-base
E2E_IMAGE ?= go-fiber-test-runner
E2E_FORCE_BUILD ?= 0
E2E_BASE_FORCE_BUILD ?= 0

.PHONY: e2e e2e-base e2e-build e2e-fast e2e-strict e2e-prune e2e-docker-clean

# E2E Docker harness (infra/)
e2e: e2e-strict

e2e-base:
	@if [ "$(E2E_BASE_FORCE_BUILD)" = "1" ] || ! docker image inspect "$(E2E_BASE_IMAGE)" >/dev/null 2>&1; then \
		echo "Building $(E2E_BASE_IMAGE)"; \
		docker build -f infra/Dockerfile.test.base -t "$(E2E_BASE_IMAGE)" .; \
	else \
		echo "Reusing $(E2E_BASE_IMAGE)"; \
	fi

e2e-build: e2e-base
	@if [ "$(E2E_FORCE_BUILD)" = "1" ] || ! docker image inspect "$(E2E_IMAGE)" >/dev/null 2>&1; then \
		echo "Building $(E2E_IMAGE)"; \
		docker build -f infra/Dockerfile.test --build-arg E2E_BASE_IMAGE="$(E2E_BASE_IMAGE)" -t "$(E2E_IMAGE)" .; \
	else \
		echo "Reusing $(E2E_IMAGE)"; \
	fi

e2e-fast: e2e-build
	@E2E_IMAGE="$(E2E_IMAGE)" \
		WORKTREE_ID="$(WORKTREE_ID)" \
		E2E_VERBOSE="$(V)" \
		docker compose -f infra/docker-compose.test.yml --project-directory . run --rm -T test-runner

e2e-strict: e2e-build
	@E2E_IMAGE="$(E2E_IMAGE)" \
		WORKTREE_ID="$(WORKTREE_ID)" \
		E2E_VERBOSE="$(V)" \
		docker compose -f infra/docker-compose.test.yml -f infra/docker-compose.test.strict.yml --project-directory . run --rm -T test-runner

e2e-prune:
	@E2E_IMAGE="$(E2E_IMAGE)" \
		WORKTREE_ID="$(WORKTREE_ID)" \
		docker compose -f infra/docker-compose.test.yml --project-directory . down -v 2>/dev/null || true
	rm -rf e2e-report

e2e-docker-clean:
	@docker image rm -f "$(E2E_IMAGE)" >/dev/null 2>&1 || true
	@docker volume rm "go-fiber-mod-$(WORKTREE_ID)" "go-fiber-build-$(WORKTREE_ID)" >/dev/null 2>&1 || true
