package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/connection"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/message"
	"github.com/khanhnguyen02311/EchoChat-WS/components/services/rabbitmq"
	"github.com/khanhnguyen02311/EchoChat-WS/configurations"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

var (
	SuccessfulConnectionCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "echo_ws",
		Name:      "successful_websocket_connections_total",
		Help:      "Total number of successful WebSocket connections",
	})
	MessageReceivedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "echo_ws",
		Name:      "messages_received_total",
		Help:      "Total number of messages received, separated by type",
	}, []string{"type"})
	MessageSentCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "echo_ws",
		Name:      "messages_sent_total",
		Help:      "Total number of messages sent, separated by type",
	}, []string{"type"})
	// Custom prometheus metrics

	e *echo.Echo
	m *manager.ConnectionManager
)

func handleMessage(c echo.Context, conn *connection.WSConnection, msg *message.InputMessage) {
	err := m.ProcessInputMessage(conn, msg)
	if err != nil {
		c.Logger().Error(err)
		_ = conn.WriteJSONMessage(message.NewOutputMessage(
			message.MsgTypeResponse,
			message.MsgStatusError,
			err.Error()))
		MessageSentCounter.WithLabelValues(message.MsgTypeResponse + "-" + message.MsgStatusError).Inc()
		return
	}
	_ = conn.WriteJSONMessage(message.NewOutputMessage(
		message.MsgTypeResponse,
		message.MsgStatusSuccess,
		""))
	MessageSentCounter.WithLabelValues(message.MsgTypeResponse + "-" + message.MsgStatusSuccess).Inc()
}

func initWS(c echo.Context) error {
	// Validate client and create connection
	conn, err := m.ValidateAndAddConnection(c.Response().Writer, c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return c.String(404, err.Error())
	}
	SuccessfulConnectionCounter.Inc()
	defer m.RemoveConnection(conn)
	for {
		msg, err := conn.ReadJSONMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				c.Logger().Info("Connection closed: ", err.Error())
				break
			} else if websocket.IsUnexpectedCloseError(err) {
				c.Logger().Error("Unexpected close error: ", err.Error())
				break
			} else {
				_ = conn.WriteJSONMessage(message.NewOutputMessage(
					message.MsgTypeResponse,
					message.MsgStatusError,
					err.Error()))
				MessageSentCounter.WithLabelValues(message.MsgTypeResponse + "-" + message.MsgStatusError).Inc()
			}
		} else {
			go handleMessage(c, conn, msg)
		}
	}
	return nil
}

func main() {
	envFile := ".env.dev"
	switch os.Getenv("APP_STAGE") {
	case "staging":
		fmt.Println("Running in staging mode")
		envFile = ".env.staging"
	case "prod":
		fmt.Println("Running in production mode")
		envFile = ".env.prod"
	default:
		fmt.Println("Running in development mode")
	}
	// Init context and environment variables
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := configurations.InitEnv(envFile); err != nil {
		fmt.Printf("Error loading environment variables: %s\n", err.Error())
		return
	}

	// Init components
	dbSession, err := db.NewScyllaSession()
	if err != nil {
		fmt.Println("Error connecting to ScyllaDB:", err.Error())
		return
	}
	//dbSession.Test()
	e = echo.New()
	m = manager.NewConnectionManager(e, dbSession, MessageSentCounter, MessageReceivedCounter)
	rmq, err := rabbitmq.NewRMQService()
	if err != nil {
		fmt.Println("Error connecting to RabbitMQ:", err.Error())
		return
	}

	// register custom prometheus metrics
	if err := prometheus.Register(SuccessfulConnectionCounter); err != nil {
		log.Fatal(err)
	}
	if err := prometheus.Register(MessageSentCounter); err != nil {
		log.Fatal(err)
	}
	if err := prometheus.Register(MessageReceivedCounter); err != nil {
		log.Fatal(err)
	}

	// pre-load counter labels
	SuccessfulConnectionCounter.Add(0)
	MessageSentCounter.WithLabelValues(message.MsgTypeResponse + "-" + message.MsgStatusError).Add(0)
	MessageSentCounter.WithLabelValues(message.MsgTypeResponse + "-" + message.MsgStatusSuccess).Add(0)
	MessageSentCounter.WithLabelValues(message.MsgTypeNotification).Add(0)
	MessageReceivedCounter.WithLabelValues(message.MsgTypeMessageNew).Add(0)
	MessageReceivedCounter.WithLabelValues(message.MsgTypeNotificationRead).Add(0)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("echo_ws"))

	e.GET("/metrics", echoprometheus.NewHandler()) // default handler
	e.GET("/ws", initWS)

	// Start the RabbitMQ consumer and the Echo server
	var wg sync.WaitGroup
	wg.Add(1)
	go rmq.StartConsuming(ctx, m, &wg)
	go func() {
		if err := e.Start("0.0.0.0:" + configurations.APP_PORT); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		// Shutdown signal received, cancel the context to signal the RabbitMQ consumer to finish
		fmt.Println("Shutting down...")
		cancel()
		rmq.Close()
	}
	// Wait for the processing goroutine to finish
	wg.Wait()
	// Gracefully shut down the Echo server
	if err := e.Shutdown(context.Background()); err != nil {
		e.Logger.Fatal(err.Error())
	}
}
