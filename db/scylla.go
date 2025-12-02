package db

import (
	"fmt"
	"go-fundraising/configs"
	"log"

	"github.com/gocql/gocql"
	"github.com/joho/godotenv"
)

var ScyllaSession *gocql.Session

func InitScylla() {
	err := godotenv.Load()

	host := configs.GetEnv("SCYLLA_HOST")
	port := configs.GetEnv("SCYLLA_PORT")
	keyspace := configs.GetEnv("SCYLLA_KEYSPACE")

	cluster := gocql.NewCluster(fmt.Sprintf("%s:%s", host, port))
	//cluster.Keyspace = "system"
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	ScyllaSession, err = cluster.CreateSession()
	if err != nil {
		log.Fatalf("can not connect ScyllaDB: %v", err)
	}
	log.Println("🔗 connected to ScyllaDB!")
	//
	//	content, err := os.ReadFile("init.cql")
	//	if err != nil {
	//		log.Fatalf("❌ cannot read init.cql: %v", err)
	//	}
	//
	//	queries := splitQueries(string(content))
	//
	//	for _, q := range queries {
	//		if strings.TrimSpace(q) == "" {
	//			continue
	//		}
	//		log.Println("▶️  Exec:", q)
	//		if err := ScyllaSession.Query(q).Exec(); err != nil {
	//			log.Println("Error:", err)
	//		}
	//	}
	//	log.Println("ScyllaDB init scripts executed")
	//
	//	appCluster := gocql.NewCluster(fmt.Sprintf("%s:%s", host, port))
	//	appCluster.Keyspace = keyspace
	//	appCluster.Consistency = gocql.Quorum
	//
	//	ScyllaSession, err = appCluster.CreateSession()
	//	if err != nil {
	//		log.Fatalf("cannot connect to ScyllaDB (app keyspace): %v", err)
	//	}
	//
	//}
	//
	//func splitQueries(cql string) []string {
	//	list := []string{}
	//	current := ""
	//
	//	for _, c := range cql {
	//		current += string(c)
	//		if c == ';' {
	//			list = append(list, current)
	//			current = ""
	//		}
	//	}
	//
	//	return list
}

func CloseScylla() {
	if ScyllaSession != nil {
		ScyllaSession.Close()
		log.Println("🔌Closed ScyllaDB")
	}
}
