package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/builderRepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/modelWriter"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	serverDsn := flag.String("dsn", "", "server dsn")
	flag.Parse()
	dbConnection := initDBConnection(*serverDsn)
	defer closeDBConnection(dbConnection)

	metadataRepo := builderRepo.MetadataRepo{Db: dbConnection}
	metadata, err := metadataRepo.GetMetadata()
	if err != nil {
		log.Fatal(err)
	}

	modelPath, err := filepath.Abs("internal/model/model.go")
	if err != nil {
		log.Fatal(err)
	}
	modelWriter := modelWriter.NewModelWriter(modelPath, metadata)
	modelWriter.Write()
}

func initDBConnection(serverDsn string) *sql.DB {
	if serverDsn == "" {
		log.Fatal("dsn flag must be set")
	}

	dbConnection, err := builderRepo.SetUpDB(serverDsn)
	if err != nil {
		log.Fatal(err)
	}
	return dbConnection
}

func closeDBConnection(connection *sql.DB) {
	err := connection.Close()
	if err != nil {
		log.Fatal(err)
	}
}
