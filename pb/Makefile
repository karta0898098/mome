# Find all .proto files.
PROTO_FILES := $(wildcard ./*/*.proto)


proto:
	$(foreach file,$(PROTO_FILES), protoc --go_out=. \
 										  --go_opt=paths=source_relative \
 										  --go-grpc_out=require_unimplemented_servers=false:. \
 										  --go-grpc_opt=paths=source_relative $(file);)
