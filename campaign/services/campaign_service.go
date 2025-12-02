package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-fundraising/campaign/models"
	"go-fundraising/db"
	"go-fundraising/worker"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type CampaignService struct{}

type SearchResult struct {
	Total int64            `json:"total"`
	Data  []map[string]any `json:"data"`
}

func (s *CampaignService) CreateCampaign(ctx context.Context, campaign models.Campaign) (models.Campaign, error) {

	stmt, names := qb.Insert(models.CampaignTable.Name).
		Columns(models.CampaignTable.Columns...).
		ToCql()

	err := gocqlx.Query(db.ScyllaSession.Query(stmt).Consistency(gocql.One), names).
		BindStruct(campaign).
		ExecRelease()

	if err != nil {
		return models.Campaign{}, err
	} else {
		worker.EnqueueSync(campaign)
	}

	return campaign, nil
}

func (s *CampaignService) GetCampaignByID(ctx context.Context, campaign_id gocql.UUID) (models.Campaign, error) {

	var campaign models.Campaign

	stmt, names := qb.Select(models.CampaignTable.Name).Where(qb.Eq("id")).ToCql()
	q := gocqlx.Query(db.ScyllaSession.Query(stmt), names).BindMap(map[string]interface{}{
		"id": campaign_id,
	})

	if err := q.GetRelease(&campaign); err != nil {
		return models.Campaign{}, err
	}
	return campaign, nil
}

func (s *CampaignService) UpdateCampaignAmountCollected(
	ctx context.Context,
	campaignID gocql.UUID,
	amount int64,
) error {

	var curr struct {
		AmountCollected int64 `db:"amount_collected"`
	}
	stmtSel, namesSel := qb.Select(models.CampaignTable.Name).
		Columns("amount_collected").
		Where(qb.Eq("id")).Limit(1).
		ToCql()

	err := gocqlx.Query(
		db.ScyllaSession.Query(stmtSel),
		namesSel,
	).BindMap(map[string]interface{}{
		"id": campaignID,
	}).GetRelease(&curr)
	if err != nil {
		return err
	}

	newAmount := curr.AmountCollected + amount

	stmtUpd, namesUpd := qb.Update(models.CampaignTable.Name).
		Set("amount_collected").
		Where(qb.Eq("id")).
		ToCql()

	return gocqlx.Query(
		db.ScyllaSession.Query(stmtUpd),
		namesUpd,
	).BindMap(map[string]interface{}{
		"id":               campaignID,
		"amount_collected": newAmount,
	}).ExecRelease()
}

func (s CampaignService) SearchCampaign(keyword string, page, perPage int) (SearchResult, error) {
	index := "campaigns"
	from := (page - 1) * perPage

	type Query struct {
		Query any           `json:"query"`
		From  int           `json:"from"`
		Size  int           `json:"size"`
		Sort  []interface{} `json:"sort,omitempty"`
	}

	var qBody Query
	if keyword == "" {
		qBody.Query = map[string]any{"match_all": map[string]any{}}
	} else {
		qBody.Query = map[string]any{
			"multi_match": map[string]any{
				"query":  keyword,
				"fields": []string{"title", "description"},
				"type":   "best_fields",
			},
		}
	}
	qBody.From = from
	qBody.Size = perPage
	qBody.Sort = []interface{}{
		map[string]any{"created_at": map[string]string{"order": "desc"}},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(qBody); err != nil {
		return SearchResult{}, fmt.Errorf("encode query: %w", err)
	}

	res, err := db.ElasticClient.Search(
		db.ElasticClient.Search.WithContext(context.Background()),
		db.ElasticClient.Search.WithIndex(index),
		db.ElasticClient.Search.WithBody(&buf),
	)
	if err != nil {
		return SearchResult{}, fmt.Errorf("es search error: %w", err)
	}
	defer res.Body.Close()

	// Parse response
	var r struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return SearchResult{}, fmt.Errorf("decode es response: %w", err)
	}

	data := make([]map[string]any, len(r.Hits.Hits))
	for i, h := range r.Hits.Hits {
		data[i] = h.Source
	}

	return SearchResult{
		Total: r.Hits.Total.Value,
		Data:  data,
	}, nil
}

func (s CampaignService) GetCampaignByUserID(ctx context.Context, userID string, page, perPage int) (SearchResult, error) {
	index := "campaigns"
	from := (page - 1) * perPage

	qBody := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"user_id": userID,
			},
		},
		"from": from,
		"size": perPage,
		"sort": []any{
			map[string]any{
				"created_at": map[string]any{
					"order": "desc",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(qBody); err != nil {
		return SearchResult{}, fmt.Errorf("encode query: %w", err)
	}

	res, err := db.ElasticClient.Search(
		db.ElasticClient.Search.WithIndex(index),
		db.ElasticClient.Search.WithBody(&buf),
	)
	if err != nil {
		return SearchResult{}, fmt.Errorf("es search error: %w", err)
	}
	defer res.Body.Close()

	var r struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return SearchResult{}, fmt.Errorf("decode es response: %w", err)
	}

	data := make([]map[string]any, len(r.Hits.Hits))
	for i, hit := range r.Hits.Hits {
		data[i] = hit.Source
	}

	return SearchResult{
		Total: r.Hits.Total.Value,
		Data:  data,
	}, nil
}
