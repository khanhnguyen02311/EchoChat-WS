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
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

var (
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
		return
	}
	_ = conn.WriteJSONMessage(message.NewOutputMessage(
		message.MsgTypeResponse,
		message.MsgStatusSuccess,
		"Message sent"))
}

func initWS(c echo.Context) error {
	// Validate client and create connection
	conn, err := m.ValidateAndAddConnection(c.Response().Writer, c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return c.String(404, err.Error())
	}
	m.GetAllConnections()
	defer m.RemoveConnection(conn)
	for {
		msg, err := conn.ReadJSONMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseNormalClosure) {
				c.Logger().Error(err)
				break
			} else {
				_ = conn.WriteJSONMessage(message.NewOutputMessage(
					message.MsgTypeResponse,
					message.MsgStatusError,
					err.Error()))
			}
		} else {
			go handleMessage(c, conn, msg)
		}
	}
	return nil
}

func main() {
	// Init context and environment variables
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := configurations.InitEnv(".env.dev"); err != nil {
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
	m = manager.NewConnectionManager(e, dbSession)
	rmq, err := rabbitmq.NewRMQService()
	if err != nil {
		return
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("myapp"))

	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/ws", initWS)

	// Start the RabbitMQ consumer and the Echo server
	var wg sync.WaitGroup
	wg.Add(1)
	go rmq.StartConsuming(ctx, m, &wg)
	go func() {
		if err := e.Start(":1323"); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

//type Server struct {
//	conns map[*websocket.Conn]bool
//}
//
//func NewServer() *Server {
//	return &Server{
//		conns: make(map[*websocket.Conn]bool),
//	}
//}
//
//func (s *Server) handleWS(ws *websocket.Conn) {
//	fmt.Println("New incoming connection from client:", ws.RemoteAddr())
//	s.conns[ws] = true
//	s.readLoop(ws)
//}
//
//func (s *Server) readLoop(ws *websocket.Conn) {
//	buf := make([]byte, 1024)
//	for {
//		n, err := ws.Read(buf)
//		if err != nil {
//			if err == io.EOF {
//				break
//			}
//			fmt.Println("Read error:", err)
//			continue
//		}
//		msg := buf[:n]
//		fmt.Println(string(msg))
//		ws.Write([]byte("Thanks for the message"))
//	}
//}
//
//func main() {
//	server := NewServer()
//	http.Handle("/ws", websocket.Handler(server.handleWS))
//	err := http.ListenAndServe(":3000", nil)
//	if err != nil {
//		fmt.Println("Server error:", err)
//	}
//}
