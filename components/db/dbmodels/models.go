package dbmodels

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/table"
	"time"
)

var (
	participantByGroupMetadata = table.Metadata{
		Name:    "participant_by_group",
		Columns: []string{"group_id", "time_created", "accountinfo_id", "notify", "role"},
		PartKey: []string{"group_id"},
		SortKey: []string{"time_created", "accountinfo_id"},
	}
	messageByGroupMetadata = table.Metadata{
		Name:    "message_by_group",
		Columns: []string{"accountinfo_id", "time_created", "group_id", "content", "type", "accountinfo_name", "group_name"},
		PartKey: []string{"accountinfo_id"},
		SortKey: []string{"time_created", "group_id"},
	}
	messageByAccountMetadata = table.Metadata{
		Name:    "message_by_group",
		Columns: []string{"group_id", "time_created", "accountinfo_id", "content", "type", "accountinfo_name", "group_name"},
		PartKey: []string{"group_id"},
		SortKey: []string{"time_created", "accountinfo_id"},
	}
	notificationMetadata = table.Metadata{
		Name:    "notification",
		Columns: []string{"accountinfo_id", "type", "time_created", "group_id", "accountinfo_id_sender", "content"},
		PartKey: []string{"accountinfo_id"},
		SortKey: []string{"type", "time_created", "group_id"},
	}
	notificationSeenMetadata = table.Metadata{
		Name:    "notification",
		Columns: []string{"accountinfo_id", "type", "group_id", "time_created", "content", "time_seen"},
		PartKey: []string{"accountinfo_id"},
		SortKey: []string{"type", "group_id", "time_created"},
	}
)

// ScyllaDBTables provides metadata for query builder only, not used for creating tables
type ScyllaDBTables struct {
	NotificationTable       *table.Table
	NotificationSeenTable   *table.Table
	MessageByGroupTable     *table.Table
	MessageByAccountTable   *table.Table
	ParticipantByGroupTable *table.Table
}

func InitScyllaDBTables() *ScyllaDBTables {
	return &ScyllaDBTables{
		NotificationTable:       table.New(notificationMetadata),
		NotificationSeenTable:   table.New(notificationSeenMetadata),
		MessageByGroupTable:     table.New(messageByGroupMetadata),
		MessageByAccountTable:   table.New(messageByAccountMetadata),
		ParticipantByGroupTable: table.New(participantByGroupMetadata),
	}
}

type Notification struct {
	AccountinfoID       int        `db:"accountinfo_id" json:"accountinfo_id"`
	Type                string     `db:"type" json:"type"`
	TimeCreated         time.Time  `db:"time_created" json:"time_created"`
	GroupID             gocql.UUID `db:"group_id" json:"group_id"`
	AccountinfoIDSender int        `db:"accountinfo_id_sender" json:"accountinfo_id_sender"`
	Content             string     `db:"content" json:"content"`
}

type NotificationSeen struct {
	AccountinfoID int        `db:"accountinfo_id" json:"accountinfo_id"`
	Type          string     `db:"type" json:"type"`
	GroupID       gocql.UUID `db:"group_id" json:"group_id"`
	TimeCreated   time.Time  `db:"time_created" json:"time_created"`
	TimeSeen      time.Time  `db:"time_seen" json:"time_seen"`
	Content       string     `db:"content" json:"content"`
}

type Participant struct {
	GroupID       gocql.UUID `db:"group_id"`
	TimeCreated   time.Time  `db:"time_created"`
	AccountinfoID int        `db:"accountinfo_id"`
	Notify        bool       `db:"notify"`
	Role          string     `db:"role"`
}

type Message struct {
	GroupID         gocql.UUID `db:"group_id" json:"group_id"`
	TimeCreated     time.Time  `db:"time_created" json:"time_created"`
	AccountinfoID   int        `db:"accountinfo_id" json:"accountinfo_id"`
	Content         string     `db:"content" json:"content"`
	Type            string     `db:"type" json:"type"`
	AccountinfoName string     `db:"accountinfo_name" json:"accountinfo_name"`
	GroupName       string     `db:"group_name" json:"group_name"`
}
