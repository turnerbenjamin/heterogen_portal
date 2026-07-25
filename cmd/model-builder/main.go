package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/builderRepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/modelWriter"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	dbConnection := initDBConnection(ctx)
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

func initDBConnection(ctx context.Context) *sql.DB {
	serverDsn, ok := os.LookupEnv("SQL_SERVER_DSN")
	if !ok {
		log.Fatal("unable to access SQL_SERVER_DSN env variable")
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
