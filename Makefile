build:
	go build -o bin/echochatws

run: build
	./bin/echochatws

proto:
	protoc --go_out=plugins=grpc:. \
	--go_opt=paths=source_relative \
	--go_grpc_out=paths=source_relative:. \
	proto/echochat.proto

.PHONY: proto


