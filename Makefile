
generate-grpc-stubs:
	protoc --go_out=api/proto/stubs --go-grpc_out=api/proto/stubs \
	  --go_opt=paths=source_relative \
	  --go-grpc_opt=paths=source_relative \
	  -I api/proto api/proto/notification.proto
