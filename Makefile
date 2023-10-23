.PHONY: build test lint fmt clean start
.DEFAULT_GOAL := all

export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
CGO_ENABLED=0

VAULT_IMAGE ?= quay.io/app-sre/vault
VAULT_VERSION ?= 1.12.7
VAULT_PORT ?= 8200

CONTAINER_ENGINE ?= $(shell which podman >/dev/null 2>&1 && echo podman || echo docker)

all: build

build:
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o vault/plugins/oidc-access-token cmd/oidc-access-token/main.go

test: build
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) go test ./...

fmt:
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) go fmt $$(go list ./...)

lint:
	$(CONTAINER_ENGINE) run --rm -w app -v "$(PWD):/app" --workdir=/app \
		quay.io/app-sre/golangci-lint:v$(shell cat .golangciversion) \
		golangci-lint run --timeout 15m

start: GOOS=linux GOARCH=amd64
start: build
	# Start a Vault server on port $VAULT_PORT and register the plugin on path oidc/
	$(CONTAINER_ENGINE) run --rm \
		-e VAULT_ADDR="http://127.0.0.1:8200" \
		-p $(VAULT_PORT):8200 \
		-v "$(PWD)/vault/plugins:/plugins" \
		$(VAULT_IMAGE):$(VAULT_VERSION) \
		sh -c "vault server -dev -dev-listen-address=0.0.0.0:8200 -dev-root-token-id=root -dev-plugin-dir=/plugins & sleep 5 && vault login root && vault secrets enable -path=oidc oidc-access-token && sleep infinity"

clean:
	rm -f ./vault/plugins/oidc-access-token
