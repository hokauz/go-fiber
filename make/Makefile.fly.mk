FLY_CONFIG ?= fly.toml
ENV_PROD_FILE ?= .env.prod

.PHONY: fly-validate deploy secret-prd

fly-validate:
	@fly config validate --config "$(FLY_CONFIG)"

deploy: fly-validate
	@fly deploy --config "$(FLY_CONFIG)"

secret-prd:
	@if [ ! -f "$(ENV_PROD_FILE)" ]; then \
		echo "$(ENV_PROD_FILE) not found"; \
		exit 1; \
	fi
	@fly secrets import < "$(ENV_PROD_FILE)"
