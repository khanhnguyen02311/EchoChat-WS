package handler

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
	"strconv"
)

type IParticipantHandler interface {
	CheckJoinedParticipant(accountinfoID int, groupID gocql.UUID) (*dbmodels.Participant, error)
	GetAllParticipantsFromGroup(groupID gocql.UUID) ([]int, error)
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
	queryInput := []string{groupID.String(), strconv.Itoa(accountinfoID)}
	participant := dbmodels.Participant{}
	err := h.db.Session.Session.Query("SELECT * FROM participant_by_group WHERE group_id = ? AND accountinfo_id = ?", queryInput).Scan(&participant)
	if err != nil {
		fmt.Println("An error occurred while checking participant", err.Error())
		return nil, err
	}
	return &participant, nil
}

func (h ParticipantHandler) GetAllParticipantsFromGroup(groupID gocql.UUID) ([]int, error) {
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
