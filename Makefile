build.server:
	@go build -o cmd/server/server cmd/server/*.go

build.agent:
	@go build -o cmd/agent/agent cmd/agent/*.go
