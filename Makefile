build:
	go build -o bin/echochatws

run: build
	./bin/echochatws

proto:
	protoc --go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	proto/EchoChat.proto

.PHONY: proto


