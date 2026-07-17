package main

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/builderRepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/model-builder/modelWriter"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
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
	dotenvPath, err := filepath.Abs("cmd/.env")
	if err != nil {
		log.Fatal(err)
	}
	appSettings, err := etc.GetAppSettings(ctx, dotenvPath)
	if err != nil {
		log.Fatal(err)
	}

	dbConnection, err := db.SetUpDB(appSettings.SqlServerDsn)
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
