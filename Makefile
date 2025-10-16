build.server:
	@go build -o cmd/server/server cmd/server/*.go

build.agent:
	@go build -o cmd/agent/agent cmd/agent/*.go

run.server:
	@go run ./cmd/server/...

run.agent:
	@go run ./cmd/agent/...

test:
	@go test -short -race -timeout 30s -coverprofile=cover.out ./... 

cover:
	@go tool cover -func=cover.out  