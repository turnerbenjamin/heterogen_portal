package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/modelwriter"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

type config struct {
	serverDsn       string
	modelSchemaName string
	enumSchemaName  string
	useCache        bool
}

func main() {
	ctx := context.Background()
	defer ctx.Done()

	config := readConfig()

	cachePath, err := filepath.Abs("cmd/modelbuilder/metadata_cache.json")
	if err != nil {
		log.Fatal(err)
	}

	var metadata *builderrepo.DatabaseSchema = nil
	if config.useCache {
		metadata, err = readFromCache(cachePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if metadata == nil {
		dbConnection := initDBConnection(config.serverDsn)
		defer closeDBConnection(dbConnection)

		metadataRepo := builderrepo.MetadataRepo{Db: dbConnection}
		metadata, err = metadataRepo.GetModelMetadata(
			config.modelSchemaName,
			config.enumSchemaName,
		)
		if err != nil {
			log.Fatal(err)
		}

		err := writeToCache(cachePath, metadata)
		if err != nil {
			log.Fatal(err)
		}
	}

	schemaData, err := schemamodel.BuildModel(metadata)
	if err != nil {
		log.Fatal(err)
	}

	tableTypesPath, err := filepath.Abs("internal/model/tableTypes.go")
	if err != nil {
		log.Fatal(err)
	}
	tableTypesWriter := modelwriter.NewTableTypesWriter(tableTypesPath, metadata)
	tableTypesWriter.Write()

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

	queryModelWriter := modelwriter.NewQueryModelWriter(queryModelPath, schemaData)
	err = queryModelWriter.Write()
	if err != nil {
		log.Fatal(err)
	}

	constantsPath, err := filepath.Abs("internal/model/constants.go")
	if err != nil {
		log.Fatal(err)
	}
	constantsWriter := modelwriter.NewConstantsWriter(constantsPath, schemaData)
	constantsWriter.Write()
}

func readConfig() config {
	serverDsn := flag.String("dsn", "", "server dsn")
	modelSchemaName := flag.String("model-schema-name", "", "model schema name")
	enumSchemaName := flag.String("enum-schema-name", "", "enum schema name")
	useCache := flag.Bool("use-cache", false, "use cache")

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
		useCache:        *useCache,
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

func writeToCache(cachePath string, metadata *builderrepo.DatabaseSchema) (err error) {
	jsonData, err := json.Marshal(metadata)

	if err != nil {
		return err
	}

	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	_, err = f.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

func readFromCache(cachePath string) (*builderrepo.DatabaseSchema, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out *builderrepo.DatabaseSchema
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
