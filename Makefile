.PHONY: build.server
build.server:
	@go build -o cmd/server/server cmd/server/*.go

.PHONY: build.agent
build.agent:
	@go build -o cmd/agent/agent cmd/agent/*.go

.PHONY: run.server
run.server:
	@go run ./cmd/server/...

.PHONY: run.agent
run.agent:
	@go run ./cmd/agent/...

.PHONY: test
test:
	@go test -short -race -timeout 30s -coverprofile=cover.out ./... 

.PHONY: cover
cover:
	@go tool cover -func=cover.out  

.PHONY: migration.sql
migration.sql:
	@goose -dir ./migrations create $(name) sql 

.PHONY: gofmt
gofmt:
	@gofmt -w .

.PHONY: goimports
goimports:
	@goimports -w -local github.com/htrandev/metrics .

generate.keys:
	@go run ./cmd/keygen...