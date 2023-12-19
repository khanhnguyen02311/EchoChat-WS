package connection

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager/message"
	"golang.org/x/exp/slices"
)

// WSConnection wraps a websocket connection.
//
// It provides methods for reading and writing JSON messages, and provides support handler for other use cases.
type WSConnection struct {
	Conn       *websocket.Conn
	ClientID   int
	ClientName string
	//ActiveGroup *uuid.UUID
}

// TODO: add support for active group

func (c *WSConnection) ReadJSONMessage() (*message.InputMessage, error) {
	msg := &message.InputMessage{}
	err := c.Conn.ReadJSON(&msg)
	if err != nil {
		//fmt.Printf("Error reading message from client %d: %s\n", c.ClientID, err.Error())
		return nil, err
	}
	// validate message contents
	switch msg.Type {
	case message.MsgTypeMessageNew:
		if msg.Data == nil || !slices.Contains(dbmodels.DBMessageType[:2], msg.Data.Type) || msg.Data.Content == "" {
			return nil, fmt.Errorf("invalid required fields for new message (group_id, content, type)")
		}
	case message.MsgTypeNotificationRead:
		if msg.Data == nil || !slices.Contains(dbmodels.DBNotificationType, msg.Data.Type) {
			return nil, fmt.Errorf("invalid required fields for marking read notification (group_id, type)")
		}
	default:
		return nil, fmt.Errorf("invalid message type (%q for sending new messages or %q for marking messages as seen)", message.MsgTypeMessageNew, message.MsgTypeNotificationRead)
	}
	fmt.Printf("Received message from client %d\n", c.ClientID)
	return msg, nil
}

func (c *WSConnection) WriteJSONMessage(msg *message.OutputMessage) error {
	return c.Conn.WriteJSON(msg)
}
