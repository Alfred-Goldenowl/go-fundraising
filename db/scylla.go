package db

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
)

var ScyllaSession *gocql.Session

func InitScylla() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Not found file .env")
	}

	host := os.Getenv("SCYLLA_HOST")
	port := os.Getenv("SCYLLA_PORT")
	keyspace := os.Getenv("SCYLLA_KEYSPACE")

	cluster := gocql.NewCluster(fmt.Sprintf("%s:%s", host, port))
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	ScyllaSession, err = cluster.CreateSession()
	if err != nil {
		log.Fatalf("can not connect ScyllaDB: %v", err)
	}
	log.Println("🔗 connected to ScyllaDB!")
}

func CloseScylla() {
	if ScyllaSession != nil {
		ScyllaSession.Close()
		log.Println("🔌Closed ScyllaDB")
	}
}
