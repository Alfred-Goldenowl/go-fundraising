package models

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

type PaymentHistory struct {
	ID         gocql.UUID `db:"id"`
	CampaignID gocql.UUID `db:"campaign_id"`
	Username   string     `db:"username"`
	UserID     gocql.UUID `db:"user_id"`
	CreatedAt  time.Time  `db:"created_at"`
	CheckoutID string     `db:"checkout_id"`
	Amount     int64      `db:"amount"`
}

var PaymentHistoryTable = table.Metadata{
	Name:    "payment_history",
	Columns: []string{"id", "campaign_id", "username", "user_id", "created_at", "checkout_id", "amount"},
	PartKey: []string{"campaign_id"},
}
