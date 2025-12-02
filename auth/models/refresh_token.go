package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type RefreshToken struct {
	UserId    gocql.UUID `db:"user_id"`
	Token     string     `db:"refresh_token"`
	ExpiresAt time.Time  `db:"expires_at"`
}

var RefreshTokenTable = table.Metadata{
	Name:    "refresh_tokens",
	Columns: []string{"user_id", "refresh_token", "expires_at"},
	PartKey: []string{"refresh_token"},
}
