package handler

import (
	"fmt"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
)

type INotificationHandler interface {
	AddNotification(notification *dbmodels.Notification) error
	AddNotificationSeen(notificationSeen *dbmodels.NotificationSeen) error
}

type NotificationHandler struct {
	db *db.ScyllaDB
}

func NewNotificationHandler(db *db.ScyllaDB) *NotificationHandler {
	return &NotificationHandler{
		db: db,
	}
}

func (h NotificationHandler) AddNotification(notification *dbmodels.Notification) error {
	err := h.db.Session.Query(h.db.Tables.NotificationTable.Insert()).BindStruct(notification).ExecRelease()
	if err != nil {
		fmt.Println("An error occurred while inserting Notification", err.Error())
		return err
	}
	return nil
}

func (h NotificationHandler) AddMultipleNotifications(accountinfoIDs []int, notification *dbmodels.Notification) {
	for _, id := range accountinfoIDs {
		notification.AccountinfoID = id
		err := h.db.Session.Query(h.db.Tables.NotificationTable.Insert()).BindStruct(notification).ExecRelease()
		if err != nil {
			fmt.Println("An error occurred while inserting Notification", err.Error())
		}
	}
}

func (h NotificationHandler) AddNotificationSeen(notificationSeen *dbmodels.NotificationSeen) error {
	err := h.db.Session.Query(h.db.Tables.NotificationSeenTable.Insert()).BindStruct(notificationSeen).ExecRelease()
	if err != nil {
		fmt.Println("An error occurred while inserting NotificationSeen", err.Error())
		return err
	}
	return nil
}
