package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type Comment struct {
	ID         gocql.UUID `db:"id"`
	CampaignID gocql.UUID `db:"campaign_id"`
	UserID     gocql.UUID `db:"user_id"`
	Username   string     `db:"username"`
	Content    string     `db:"content"`
	CreatedAt  time.Time  `db:"created_at"`
}

var CommentTable = table.Metadata{
	Name:    "comments",
	Columns: []string{"id", "campaign_id", "user_id", "content", "created_at", "username"},
	PartKey: []string{"id", "campaign_id"},
}
