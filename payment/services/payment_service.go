package services

import (
	"context"
	"go-fundraising/db"
	"go-fundraising/payment/models"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type PaymentService struct{}

func (s *PaymentService) NewPayment(ctx context.Context, paymentHistory models.PaymentHistory) error {
	stmt, names := qb.Insert(models.PaymentHistoryTable.Name).
		Columns(models.PaymentHistoryTable.Columns...).
		ToCql()

	return gocqlx.Query(db.ScyllaSession.Query(stmt).Consistency(gocql.One), names).
		BindStruct(paymentHistory).
		ExecRelease()

}

func (s *PaymentService) GetPaymentsByCampaignID(
	ctx context.Context,
	campaignID gocql.UUID,
) ([]models.PaymentHistory, error) {

	var results []models.PaymentHistory

	stmt, names := qb.Select(models.PaymentHistoryTable.Name).
		Where(qb.Eq("campaign_id")).
		ToCql()

	q := gocqlx.Query(
		db.ScyllaSession.Query(stmt).Consistency(gocql.One),
		names,
	).BindMap(map[string]interface{}{
		"campaign_id": campaignID,
	})

	if err := q.SelectRelease(&results); err != nil {
		return nil, err
	}

	return results, nil
}
