package main

import (
	"errors"
	"github.com/gorilla/websocket"
	utils "github.com/khanhnguyen02311/EchoChat-WS/utils"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
)

var (
	e = echo.New()
	m = utils.NewConnectionManager(e)
)

func handleMessage(c echo.Context, message *utils.InputMessage) {
	err := m.SendToClient(message.Recipient, message.Type, utils.MsgStatusSuccess, message.Content)
	if err != nil {
		c.Logger().Error(err)
	}
}

func initWS(c echo.Context) error {
	// Validate client and create connection
	conn, err := m.ValidateAndAddConnection(c.Response().Writer, c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return c.String(404, err.Error())
	}
	defer m.RemoveConnection(conn)
	for {
		message, err := conn.ReadJSONMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Logger().Error(err)
			}
			break
		}
		go handleMessage(c, message)
	}
	return nil
}

func main() {
	//e := echo.New()
	//m := manager.NewConnectionManager(e)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("myapp"))

	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/ws", initWS)

	if err := e.Start(":1323"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
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
