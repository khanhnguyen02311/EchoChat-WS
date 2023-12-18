FROM golang:1.21 as builder

WORKDIR /app

RUN apt-get update &&\
    apt-get install --no-install-recommends -y \
      protobuf-compiler

COPY go.mod go.sum ./

RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

COPY . .

RUN protoc --go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	./components/services/proto/EchoChat.proto


RUN go build -o bin/echochatws

FROM golang:1.21 as runner

WORKDIR /app

COPY --from=builder /app/bin/echochatws ./bin/echochatws
COPY --from=builder /app/.env.dev ./.env.dev
COPY --from=builder /app/.env.staging ./.env.staging

STOPSIGNAL SIGQUIT

CMD CGO_ENABLED=0 GOOS=linux ./bin/echochatws
