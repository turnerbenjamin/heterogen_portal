package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/modelwriter"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	serverDsn := flag.String("dsn", "", "server dsn")
	schemaName := flag.String("schema-name", "", "schema name")
	flag.Parse()
	if *schemaName == "" {
		log.Fatal("schema-name flag must be set")
	}

	dbConnection := initDBConnection(*serverDsn)
	defer closeDBConnection(dbConnection)

	metadataRepo := builderrepo.MetadataRepo{Db: dbConnection}
	metadata, err := metadataRepo.GetMetadata(*schemaName)
	if err != nil {
		log.Fatal(err)
	}

	queryModelPath, err := filepath.Abs("internal/model/queryModel.go")
	if err != nil {
		log.Fatal(err)
	}
	queryModelWriter := modelwriter.NewQueryModelWriter(queryModelPath, metadata)
	queryModelWriter.Write()

	constantsPath, err := filepath.Abs("internal/model/constants.go")
	if err != nil {
		log.Fatal(err)
	}
	constantsWriter := modelwriter.NewConstantsWriter(constantsPath, metadata)
	constantsWriter.Write()
}

func initDBConnection(serverDsn string) *sql.DB {
	if serverDsn == "" {
		log.Fatal("dsn flag must be set")
	}

	dbConnection, err := builderrepo.SetUpDB(serverDsn)
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
