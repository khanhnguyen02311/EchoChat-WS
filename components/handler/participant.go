package handler

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
)

type IParticipantHandler interface {
	CheckJoinedParticipant(accountinfoID int, groupID gocql.UUID) (*dbmodels.Participant, error)
	GetAllParticipantIDsFromGroup(groupID gocql.UUID) ([]int, error)
}

type ParticipantHandler struct {
	db *db.ScyllaDB
}

func NewParticipantHandler(db *db.ScyllaDB) *ParticipantHandler {
	return &ParticipantHandler{
		db: db,
	}
}

func (h ParticipantHandler) CheckJoinedParticipant(accountinfoID int, groupID gocql.UUID) (*dbmodels.Participant, error) {
	participant := dbmodels.Participant{}
	err := h.db.Session.Session.Query("SELECT * FROM participant_by_account WHERE accountinfo_id = ? AND group_id = ?", accountinfoID, groupID).Scan(
		&participant.AccountinfoID, &participant.GroupID, &participant.TimeCreated, &participant.Notify, &participant.Role)
	if err != nil {
		fmt.Println("An error occurred while checking participant", err.Error())
		return nil, err
	}
	return &participant, nil
}

func (h ParticipantHandler) GetAllParticipantIDsFromGroup(groupID gocql.UUID) ([]int, error) {
	var ID int
	var accountinfoIDs []int
	// use gocql instead of gocqlx
	iter := h.db.Session.Session.Query("SELECT accountinfo_id FROM participant_by_group WHERE group_id = ?", groupID).Iter()
	for iter.Scan(&ID) {
		accountinfoIDs = append(accountinfoIDs, ID)
	}
	if err := iter.Close(); err != nil {
		fmt.Println("An error occurred while getting all participants", err.Error())
		return nil, err
	}
	return accountinfoIDs, nil
}
