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

type config struct {
	serverDsn       string
	modelSchemaName string
	enumSchemaName  string
}

func main() {
	ctx := context.Background()
	defer ctx.Done()

	config := readConfig()

	dbConnection := initDBConnection(config.serverDsn)
	defer closeDBConnection(dbConnection)

	metadataRepo := builderrepo.MetadataRepo{Db: dbConnection}
	metadata, err := metadataRepo.GetModelMetadata(
		config.modelSchemaName,
		config.enumSchemaName,
	)
	if err != nil {
		log.Fatal(err)
	}

	enumPath, err := filepath.Abs("internal/model/enums.go")
	if err != nil {
		log.Fatal(err)
	}
	enumWriter := modelwriter.NewEnumWriter(enumPath, metadata.Enums)
	enumWriter.Write()

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

func readConfig() config {
	serverDsn := flag.String("dsn", "", "server dsn")
	modelSchemaName := flag.String("model-schema-name", "", "model schema name")
	enumSchemaName := flag.String("enum-schema-name", "", "enum schema name")

	flag.Parse()

	if *serverDsn == "" {
		log.Fatal("dsn flag must be set")
	}

	if *modelSchemaName == "" {
		log.Fatal("model-schema-name flag must be set")
	}

	if *enumSchemaName == "" {
		log.Fatal("enum-schema-name flag must be set")
	}
	return config{
		serverDsn:       *serverDsn,
		modelSchemaName: *modelSchemaName,
		enumSchemaName:  *enumSchemaName,
	}
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
