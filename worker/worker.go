package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"go-fundraising/campaign/models"
	"go-fundraising/db"
	"log"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var syncJobs chan models.Campaign

func InitSyncWorkers(workerCount int) {
	syncJobs = make(chan models.Campaign, 1000)

	for i := 0; i < workerCount; i++ {
		go syncWorker(i)
	}

	log.Printf("🚀 Started %d ES sync workers\n", workerCount)
}

func syncWorker(id int) {
	for campaign := range syncJobs {
		syncToES(id, campaign)
	}
}

func syncToES(workerID int, campaign models.Campaign) {
	index := "campaigns"

	body := map[string]interface{}{
		"id":          campaign.ID.String(),
		"username":    campaign.Username,
		"user_id":     campaign.UserID,
		"title":       campaign.Title,
		"description": campaign.Description,
		"image":       campaign.Image,
		"deadline":    campaign.Deadline,
		"created_at":  campaign.CreatedAt,
	}

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: campaign.ID.String(),
		Body:       &buf,
		Refresh:    "false",
	}

	res, err := req.Do(context.Background(), db.ElasticClient)
	if err != nil {
		log.Printf("❌ Worker %d error indexing: %v\n", workerID, err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("❌ Worker %d ES index error: %s\n", workerID, res.Status())
	} else {
		log.Printf("✔️ Worker %d synced campaign %s\n", workerID, campaign.ID.String())
	}
}

func EnqueueSync(campaign models.Campaign) {
	select {
	case syncJobs <- campaign:
	default:
		log.Println("⚠️ Sync queue full, dropping job:", campaign.ID)
	}
}
