package configurations

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

var (
	PROTO_RMQ_USR        string
	PROTO_RMQ_PWD        string
	PROTO_RMQ_PORT       string
	PROTO_RMQ_QUEUE_NOTI string
	PROTO_RMQ_QUEUE_MSG  string
	PROTO_GRPC_PORT      string

	SCYLLA_HOST     string
	SCYLLA_PORT     int
	SCYLLA_KEYSPACE string
)

func InitEnv(envDirectory string) error {
	err := godotenv.Load(envDirectory)
	if err != nil {
		return err
	}

	SCYLLA_HOST = os.Getenv("SCYLLA_HOST")
	SCYLLA_PORT, _ = strconv.Atoi(os.Getenv("SCYLLA_PORT"))
	SCYLLA_KEYSPACE = os.Getenv("SCYLLA_KEYSPACE")

	PROTO_RMQ_QUEUE_NOTI = os.Getenv("PROTO_RMQ_QUEUE_NOTI")
	PROTO_RMQ_QUEUE_MSG = os.Getenv("PROTO_RMQ_QUEUE_MSG")
	PROTO_RMQ_USR = os.Getenv("PROTO_RMQ_USR")
	PROTO_RMQ_PWD = os.Getenv("PROTO_RMQ_PWD")
	PROTO_RMQ_PORT = os.Getenv("PROTO_RMQ_PORT")
	PROTO_GRPC_PORT = os.Getenv("PROTO_GRPC_PORT")

	fmt.Printf("Environment variables loaded successfully. Test variable: %s\n", SCYLLA_HOST)
	return nil
}
