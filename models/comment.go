package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type Comment struct {
	ID        gocql.UUID `db:"id"`
	PostID    string     `db:"post_id"`
	UserID    string     `db:"user_id"`
	Content   string     `db:"content"`
	CreatedAt time.Time  `db:"created_at"`
}

var CommentTable = table.Metadata{
	Name:    "comments",
	Columns: []string{"id", "post_id", "user_id", "content", "created_at"},
	PartKey: []string{"id", "post_id"},
}
