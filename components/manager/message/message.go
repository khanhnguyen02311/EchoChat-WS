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
	Type string                `json:"type"`
	Data *dbmodels.MessagePOST `json:"data"`
}

type OutputMessage struct {
	Type         string                    `json:"type"`
	Status       string                    `json:"status"`
	Message      *servicemodels.RMQMessage `json:"message"`
	Notification *dbmodels.Notification    `json:"notification"`
	Content      string                    `json:"content"`
}

// TODO: Find a way to bind item inside output message

func NewInputMessage(msgType string, data *dbmodels.MessagePOST) *InputMessage {
	return &InputMessage{
		Type: msgType,
		Data: data,
	}
}

func NewOutputMessage(msgType string, msgStatus string, msgContent string) *OutputMessage {
	return &OutputMessage{
		Type:    msgType,
		Status:  msgStatus,
		Content: msgContent,
	}
}
