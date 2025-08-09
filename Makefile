help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate-grpc-stubs: ## Generate gRPC stubs from proto files
	protoc --go_out=api/proto/stubs --go-grpc_out=api/proto/stubs \
	  --go_opt=paths=source_relative \
	  --go-grpc_opt=paths=source_relative \
	  -I api/proto api/proto/notification.proto
