package configurations

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

var (
	APP_PORT string

	PROTO_RMQ_USR        string
	PROTO_RMQ_PWD        string
	PROTO_RMQ_HOST       string
	PROTO_RMQ_PORT       string
	PROTO_RMQ_QUEUE_NOTI string
	PROTO_RMQ_QUEUE_MSG  string

	PROTO_GRPC_HOST string
	PROTO_GRPC_PORT string

	SCYLLA_HOST     string
	SCYLLA_PORT     int
	SCYLLA_KEYSPACE string
)

func InitEnv(envFile string) error {
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			return err
		}
	}
	APP_PORT = os.Getenv("APP_PORT")

	PROTO_RMQ_QUEUE_NOTI = os.Getenv("PROTO_RMQ_QUEUE_NOTI")
	PROTO_RMQ_QUEUE_MSG = os.Getenv("PROTO_RMQ_QUEUE_MSG")
	PROTO_RMQ_USR = os.Getenv("PROTO_RMQ_USR")
	PROTO_RMQ_PWD = os.Getenv("PROTO_RMQ_PWD")
	PROTO_RMQ_HOST = os.Getenv("PROTO_RMQ_HOST")
	PROTO_RMQ_PORT = os.Getenv("PROTO_RMQ_PORT")

	PROTO_GRPC_HOST = os.Getenv("PROTO_GRPC_HOST")
	PROTO_GRPC_PORT = os.Getenv("PROTO_GRPC_PORT")

	SCYLLA_HOST = os.Getenv("SCYLLA_HOST")
	SCYLLA_PORT, _ = strconv.Atoi(os.Getenv("SCYLLA_PORT"))
	SCYLLA_KEYSPACE = os.Getenv("SCYLLA_KEYSPACE")

	fmt.Printf("Environment variables loaded successfully. Application port: %s\n", APP_PORT)
	return nil
}
