package dbmodels

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/table"
	"time"
)

var (
	DBMessageType      = []string{"Message", "File", "Event", "Other"}
	DBNotificationType = []string{"GroupEvent", "GroupRequest", "Other"}

	groupMetadata = table.Metadata{
		Name:    "group",
		Columns: []string{"id", "name", "description", "visibility", "time_created"},
		PartKey: []string{"id"},
		SortKey: []string{},
	}
	//participantByGroupMetadata = table.Metadata{
	//	Name:    "participant_by_account",
	//	Columns: []string{"group_id", "time_created", "accountinfo_id", "notify", "role"},
	//	PartKey: []string{"group_id"},
	//	SortKey: []string{"time_created", "accountinfo_id"},
	//}
	//participantByAccountMetadata = table.Metadata{
	//	Name:    "participant_by_account",
	//	Columns: []string{"accountinfo_id", "group_id", "time_created", "notify", "role"},
	//	PartKey: []string{"accountinfo_id"},
	//	SortKey: []string{"group_id", "time_created"},
	//}
	messageByGroupMetadata = table.Metadata{
		Name:    "message_by_group",
		Columns: []string{"accountinfo_id", "time_created", "group_id", "content", "type", "accountinfo_name", "group_name"},
		PartKey: []string{"accountinfo_id"},
		SortKey: []string{"time_created", "group_id"},
	}
	messageByAccountMetadata = table.Metadata{
		Name:    "message_by_account",
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
		Name:    "notification_seen",
		Columns: []string{"accountinfo_id", "type", "group_id", "time_created"},
		PartKey: []string{"accountinfo_id"},
		SortKey: []string{"type", "group_id", "time_created"},
	}
)

// ScyllaDBTables provides metadata for query builder only, not used for creating tables
type ScyllaDBTables struct {
	NotificationTable     *table.Table
	NotificationSeenTable *table.Table
	MessageByGroupTable   *table.Table
	MessageByAccountTable *table.Table
	//ParticipantByAccountTable *table.Table
	//ParticipantByGroupTable   *table.Table
	GroupTable *table.Table
}

func InitScyllaDBTables() *ScyllaDBTables {
	return &ScyllaDBTables{
		NotificationTable:     table.New(notificationMetadata),
		NotificationSeenTable: table.New(notificationSeenMetadata),
		MessageByGroupTable:   table.New(messageByGroupMetadata),
		MessageByAccountTable: table.New(messageByAccountMetadata),
		//ParticipantByAccountTable: table.New(participantByAccountMetadata),
		//ParticipantByGroupTable:   table.New(participantByGroupMetadata),
		GroupTable: table.New(groupMetadata),
	}
}

type Group struct {
	ID          gocql.UUID `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Visibility  bool       `db:"visibility" json:"visibility"`
	TimeCreated time.Time  `db:"time_created" json:"time_created"`
}

type Participant struct {
	AccountinfoID int        `db:"accountinfo_id"`
	GroupID       gocql.UUID `db:"group_id"`
	TimeCreated   time.Time  `db:"time_created"`
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

type MessagePOST struct {
	GroupID       gocql.UUID `json:"group_id"`
	AccountinfoID int        `json:"accountinfo_id"`
	Content       string     `json:"content"`
	Type          string     `json:"type"`
}

type Notification struct {
	AccountinfoID       int        `db:"accountinfo_id" json:"-"`
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
}
