.PHONY: build test clean dev install

# CGO flags for SQLite FTS5 support
export CGO_CFLAGS = -DSQLITE_ENABLE_FTS5
export CGO_LDFLAGS = -lm

build:
	GOTOOLCHAIN=auto go build -o athena ./cmd/athena/

install:
	./install.sh

build-agent:
	GOTOOLCHAIN=auto go build -o athena-agent ./cmd/athena-agent/ 2>/dev/null || true

test:
	GOTOOLCHAIN=auto go test ./... -v

test-short:
	GOTOOLCHAIN=auto go test ./... -short

clean:
	rm -f athena athena-agent
	rm -rf frontend/dist

dev: build
	./athena config/athena.yaml

frontend:
	cd frontend && npm install && npx vite build

all: frontend build

# Run server with example config
run: build
	ATHENA_LLM_BASE_URL=$${ATHENA_LLM_BASE_URL:-https://api.openai.com/v1} \
	ATHENA_LLM_API_KEY=$${ATHENA_LLM_API_KEY} \
	ATHENA_LLM_MODEL=$${ATHENA_LLM_MODEL:-gpt-4o} \
	./athena
