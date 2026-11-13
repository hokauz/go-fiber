GOPATH := $(strip $(shell go env GOPATH))
GOBIN := $(strip $(shell go env GOBIN))
ifeq ($(GOBIN),)
GOBIN := $(GOPATH)/bin
endif

SWAG_PATH := $(shell command -v swag 2>/dev/null)
SWAG_BIN := $(if $(SWAG_PATH),$(SWAG_PATH),$(GOBIN)/swag)

.PHONY: ensure-swag regen-docs

ensure-swag:
	@if [ ! -x "$(SWAG_BIN)" ]; then \
		echo "swag not found. Installing..."; \
		GOBIN=$(GOBIN) go install github.com/swaggo/swag/cmd/swag@latest; \
	fi

regen-docs: ensure-swag
	@mkdir -p docs
	@$(SWAG_BIN) init -g main.go -d cmd/api,internal -o docs --parseDependency --parseInternal --outputTypes json
	@go run ./cmd/scalar -in docs/swagger.json -out docs/openapi.json
	@echo "docs/openapi.json generated"
