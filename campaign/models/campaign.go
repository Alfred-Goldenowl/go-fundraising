package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type Campaign struct {
	ID              gocql.UUID `db:"id"`
	UserID          gocql.UUID `db:"user_id"`
	Username        string     `db:"username"`
	Title           string     `db:"title"`
	Description     string     `db:"description"`
	Target          int        `db:"target"`
	AmountCollected int        `db:"amount_collected"`
	Image           string     `db:"image"`
	Deadline        time.Time  `db:"deadline"`
	CreatedAt       time.Time  `db:"created_at"`
}

var CampaignTable = table.Metadata{
	Name: "campaigns",
	Columns: []string{
		"id",
		"user_id",
		"username",
		"title",
		"description",
		"target",
		"amount_collected",
		"image",
		"deadline",
		"created_at",
	},
}
