build:
	@go build -o ./bin/vinomapi ./main.go

test:
	go test -v ./...

run: build
	@./bin/vinomapi

# Variables (can be overridden when calling `make`)
PROTO_FILE ?= ./udp/pb_encoder/*.proto  # Path to the .proto file
PROTO_PATH ?= ./udp/pb_encoder               # Path to the directory containing .proto files
OUT_PATH   ?= ./udp/pb_encoder               # Output directory for generated files

# Rule to generate Go code
genpb: 
	@echo "Generating Go code from $(PROTO_FILE)..."
	protoc -I $(PROTO_PATH) --go_out=$(OUT_PATH) $(PROTO_FILE)

