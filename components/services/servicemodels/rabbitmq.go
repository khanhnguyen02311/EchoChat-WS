package servicemodels

import (
	"github.com/gocql/gocql"
	"github.com/relvacode/iso8601"
)

type RMQMessage struct {
	GroupID         gocql.UUID    `json:"group_id"`
	TimeCreated     *iso8601.Time `json:"time_created"`
	AccountinfoID   int           `json:"accountinfo_id"`
	Type            string        `json:"type"`
	Content         string        `json:"content"`
	GroupName       string        `json:"group_name"`
	AccountinfoName string        `json:"accountinfo_name"`
}

// TODO: find ways to convert iso8601.Time to time.Time cleanly
