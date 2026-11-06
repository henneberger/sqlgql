DATABASE_URL ?= postgres://demo:demo@127.0.0.1:55432/demo?sslmode=disable
export DATABASE_URL

.PHONY: up generate test integration policy-test server web smoke down
up:
	docker compose up -d --wait
generate:
	go generate ./example
test:
	go test ./...
	npm --prefix example/web run build
integration:
	TEST_DATABASE_URL="$(DATABASE_URL)" go test -v ./example
policy-test:
	opa test example/policy
server:
	go run ./example/cmd/server
web:
	npm --prefix example/web run dev
smoke:
	npm --prefix example/web run smoke
down:
	docker compose down
