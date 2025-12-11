.PHONY: all proto server clean

all: proto server

proto:
	@echo "Generating Go code from protobuf..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		echobridge.proto

server:
	@echo "Building EchoBridge server..."
	go build -o echobridge_server main.go

clean:
	@echo "Cleaning generated files..."
	rm -f *.pb.go echobridge_server

install-deps:
	@echo "Installing dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
