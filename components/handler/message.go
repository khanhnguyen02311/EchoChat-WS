package handler

import (
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
)

type IMessageHandler interface {
	AddNewMessage(message *dbmodels.Message) error
}

type MessageHandler struct {
	db *db.ScyllaDB
}

func NewMessageHandler(db *db.ScyllaDB) *MessageHandler {
	return &MessageHandler{
		db: db,
	}
}

func (h MessageHandler) AddNewMessage(message *dbmodels.Message) error {
	err := h.db.Session.Query(h.db.Tables.MessageByGroupTable.Insert()).BindStruct(message).ExecRelease()
	if err != nil {
		return err
	}
	err = h.db.Session.Query(h.db.Tables.MessageByAccountTable.Insert()).BindStruct(message).ExecRelease()
	if err != nil {
		return err
	}
	return nil
}
