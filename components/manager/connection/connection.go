package connection

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/message"
)

// WSConnection wraps a websocket connection.
//
// It provides methods for reading and writing JSON messages, and provides support handler for other use cases.
type WSConnection struct {
	Conn     *websocket.Conn
	ClientID int
	//ActiveGroup *uuid.UUID
}

// TODO: add support for active group

func (c *WSConnection) ReadJSONMessage() (*message.InputMessage, error) {
	msg := &message.InputMessage{}
	err := c.Conn.ReadJSON(msg)
	if err != nil {
		//fmt.Printf("Error reading message from client %d: %s\n", c.ClientID, err.Error())
		return nil, err
	}
	fmt.Printf("Received message from client %d: %s\n", c.ClientID, string(msg.Content))
	return msg, nil
}

func (c *WSConnection) WriteJSONMessage(msg *message.OutputMessage) error {
	return c.Conn.WriteJSON(msg)
}
