package message

import (
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
	"github.com/khanhnguyen02311/EchoChat-WS/components/services/servicemodels"
)

const (
	MsgTypeMessage      = "message"
	MsgTypeNotification = "notification"
	MsgTypeResponse     = "response"
	MsgTypeHelp         = "help"

	MsgStatusNew     = "new"
	MsgStatusSuccess = "success"
	MsgStatusError   = "error"
	MsgStatusOther   = "other"
)

type InputMessage struct {
	Type      string `json:"type"`
	Recipient int    `json:"recipient"`
	Content   []byte `json:"content"`
}

type OutputMessage struct {
	Type         string                    `json:"type"`
	Status       string                    `json:"status"`
	Message      *servicemodels.RMQMessage `json:"message"`
	Notification *dbmodels.Notification    `json:"notification"`
	Content      []byte                    `json:"content"`
}

// TODO: Find a way to bind item inside output message

func NewInputMessage(msgType string, msgRecipient int, msgContent []byte) *InputMessage {
	return &InputMessage{
		Type:      msgType,
		Recipient: msgRecipient,
		Content:   msgContent,
	}
}

func NewOutputMessage(msgType string, msgStatus string, msgContent []byte) *OutputMessage {
	return &OutputMessage{
		Type:    msgType,
		Status:  msgStatus,
		Content: msgContent,
	}
}
