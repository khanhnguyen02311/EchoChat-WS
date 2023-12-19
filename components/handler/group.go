package handler

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
)

type IGroupHandler interface {
	GetGroupByID(groupID gocql.UUID) (*dbmodels.Group, error)
}

type GroupHandler struct {
	db *db.ScyllaDB
}

func NewGroupHandler(db *db.ScyllaDB) *GroupHandler {
	return &GroupHandler{
		db: db,
	}
}

func (h GroupHandler) GetGroupByID(groupID gocql.UUID) (*dbmodels.Group, error) {
	group := dbmodels.Group{ID: groupID}
	err := h.db.Session.Query(h.db.Tables.GroupTable.Get()).BindStruct(group).GetRelease(&group)
	if err != nil {
		fmt.Println("An error occurred while getting group", err.Error())
		return nil, err
	}
	return &group, nil
}
