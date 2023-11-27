package manager

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strconv"
)

const (
	MsgTypeMessage      = "message"
	MsgTypeNotification = "notification"
	MsgTypeResponse     = "response"
	MsgTypeHelp         = "help"

	MsgStatusSuccess = "success"
	MsgStatusError   = "error"
	MsgStatusOther   = "other"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Accepting all requests
		},
	}
)

type InputMessage struct {
	Type      string `json:"type"`
	Recipient int    `json:"recipient"`
	Content   string `json:"content"`
}
type OutputMessage struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Content string `json:"content"`
}

func NewInputMessage(msgType string, msgRecipient int, msgContent string) *InputMessage {
	return &InputMessage{
		Type:      msgType,
		Recipient: msgRecipient,
		Content:   msgContent,
	}
}

func NewOutputMessage(msgType string, msgStatus string, msgContent string) *OutputMessage {
	return &OutputMessage{
		Type:    msgType,
		Status:  msgStatus,
		Content: msgContent,
	}
}

type EchoChatConn struct {
	Conn     *websocket.Conn
	ClientID int
}

func (c *EchoChatConn) ReadJSONMessage() (*InputMessage, error) {
	message := &InputMessage{}
	err := c.Conn.ReadJSON(message)
	if err != nil {
		fmt.Printf("Error reading message from client %d: %s\n", c.ClientID, err.Error())
		return nil, err
	}
	fmt.Printf("Received message from client %d: %s\n", c.ClientID, message.Content)
	return message, nil
}

func (c *EchoChatConn) WriteJSONMessage(message *OutputMessage) error {
	return c.Conn.WriteJSON(message)
}

// A ConnectionManager handles all connections to the EchoChat server and their respective clients.
//
// It uses Echo to handle the HTTP requests/metrics/logging and custom Gorilla Websockets to handle the WebSocket connections.
// Each client ID can have multiple connections stored in slices.
type ConnectionManager struct {
	ConnsByID  map[int][]*EchoChatConn
	EchoServer *echo.Echo
}

func NewConnectionManager(e *echo.Echo) *ConnectionManager {
	return &ConnectionManager{
		ConnsByID:  make(map[int][]*EchoChatConn),
		EchoServer: e,
	}
}

func (manager *ConnectionManager) validateClient(token string) int {
	//TODO: validate the input token
	resp, err := http.Post("http://localhost:8000/auth/token/validate?token="+token, "POST", nil)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error validating token: %s\n", err.Error())
		return 0
	}
	defer resp.Body.Close()
	if body, err := io.ReadAll(resp.Body); err != nil {
		return 0
	} else {
		fmt.Printf("%s", body)
		clientID, _ := strconv.Atoi(string(body))
		return clientID
	}
}

func (manager *ConnectionManager) AddConnection(conn *websocket.Conn, clientID int) *EchoChatConn {
	c := &EchoChatConn{
		Conn:     conn,
		ClientID: clientID,
	}
	manager.ConnsByID[c.ClientID] = append(manager.ConnsByID[c.ClientID], c)
	// print all connections
	fmt.Println("Current connections:")
	for client, conns := range manager.ConnsByID {
		fmt.Printf("Client %d:\n", client)
		for _, c := range conns {
			fmt.Printf("\t%s\n", c.Conn.RemoteAddr())
		}
	}
	return c
}

func (manager *ConnectionManager) ValidateAndAddConnection(w http.ResponseWriter, r *http.Request, respHeader http.Header) (*EchoChatConn, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, errors.New("missing token")
	}
	clientID := manager.validateClient(token)
	if clientID == 0 {
		return nil, errors.New("client not found")
	}
	ws, err := upgrader.Upgrade(w, r, respHeader)
	if err != nil {
		return nil, err
	}
	return manager.AddConnection(ws, clientID), nil
}

func (manager *ConnectionManager) RemoveConnection(c *EchoChatConn) {
	defer c.Conn.Close()
	for i, conn := range manager.ConnsByID[c.ClientID] {
		if conn == c {
			if len(manager.ConnsByID[c.ClientID]) == 1 {
				delete(manager.ConnsByID, c.ClientID)
			} else {
				manager.ConnsByID[c.ClientID] = append(manager.ConnsByID[c.ClientID][:i], manager.ConnsByID[c.ClientID][i+1:]...)
			}
			break
		}
	}
}

func (manager *ConnectionManager) SendToClient(clientID int, msgType string, msgStatus string, msg string) error {
	wg := new(errgroup.Group)
	newMsg := NewOutputMessage(msgType, msgStatus, msg)
	for _, c := range manager.ConnsByID[clientID] {
		c := c // https://golang.org/doc/faq#closures_and_goroutines
		wg.Go(func() error {
			fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
			return c.WriteJSONMessage(newMsg)
		})
	}
	// handle first non-nil error if needed
	if err := wg.Wait(); err != nil {
		return err
	}
	return nil
}

func (manager *ConnectionManager) SendToClients(clientIDs []int, msgType string, msgStatus string, msg string) error {
	wg := new(errgroup.Group)
	newMsg := NewOutputMessage(msgType, msgStatus, msg)
	for client := range clientIDs {
		for _, c := range manager.ConnsByID[client] {
			c := c // https://golang.org/doc/faq#closures_and_goroutines
			wg.Go(func() error {
				fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
				return c.WriteJSONMessage(newMsg)
			})
		}
	}
	// handle first non-nil error if needed
	if err := wg.Wait(); err != nil {
		return err
	}
	return nil
}

func (manager *ConnectionManager) SendToAll(msgType string, msgStatus string, msg string) error {
	wg := new(errgroup.Group)
	newMsg := NewOutputMessage(msgType, msgStatus, msg)
	for _, conns := range manager.ConnsByID {
		for _, c := range conns {
			c := c // https://golang.org/doc/faq#closures_and_goroutines
			wg.Go(func() error {
				fmt.Printf("Sending message to client %d at %s\n", c.ClientID, c.Conn.RemoteAddr())
				return c.WriteJSONMessage(newMsg)
			})
		}
	}
	// handle first non-nil error if needed
	if err := wg.Wait(); err != nil {
		return err
	}
	return nil
}
