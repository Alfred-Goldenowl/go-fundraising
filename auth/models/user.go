package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type User struct {
	ID           gocql.UUID `db:"id"`
	Email        string     `db:"email"`
	Username     string     `db:"username"`
	PasswordHash string     `db:"password_hash"`
	CreatedAt    time.Time  `db:"created_at"`
}

var UserTable = table.Metadata{
	Name:    "users",
	Columns: []string{"id", "email", "username", "password_hash", "created_at"},
	PartKey: []string{"id"},
}
