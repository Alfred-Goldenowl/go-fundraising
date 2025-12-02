package db

import (
	"context"
	"log"
	"strings"

	"go-fundraising/configs"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/joho/godotenv"
)

var ElasticClient *elasticsearch.Client

func InitElastic() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env not found, using system env")
	}

	url := configs.GetEnv("ELASTIC_URL")

	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	ElasticClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ cannot create ElasticSearch client: %v", err)
	}

	res, err := ElasticClient.Info()
	if err != nil {
		log.Fatalf("❌ cannot ping ElasticSearch: %v", err)
	}
	defer res.Body.Close()

	log.Println("🔗 Connected to Elasticsearch:", url)

}

func splitESQueries(text string) []string {
	return strings.Split(text, "---")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func doCreateIndex(index string, query string) {
	res, err := ElasticClient.Indices.Create(index,
		ElasticClient.Indices.Create.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		log.Println("❌ Error create index:", err)
		return
	}
	defer res.Body.Close()
	log.Println("✔️ Create index result:", res.Status())
}

func doPutMapping(index string, query string) {
	req := esapi.IndicesPutMappingRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}

	res, err := req.Do(context.Background(), ElasticClient)
	if err != nil {
		log.Println("❌ PutMapping error:", err)
		return
	}
	defer res.Body.Close()

	log.Println("✔️ PutMapping:", res.Status())
}

func doPutSettings(index string, query string) {
	res, err := ElasticClient.Indices.PutSettings(strings.NewReader(query),
		ElasticClient.Indices.PutSettings.WithIndex(index),
	)
	if err != nil {
		log.Println("❌ Error put settings:", err)
		return
	}
	defer res.Body.Close()
	log.Println("✔️ Put settings:", res.Status())
}

func doBulk(query string) {
	res, err := ElasticClient.Bulk(strings.NewReader(query), ElasticClient.Bulk.WithContext(context.Background()))
	if err != nil {
		log.Println("❌ Error bulk:", err)
		return
	}
	defer res.Body.Close()
	log.Println("✔️ Bulk:", res.Status())
}
