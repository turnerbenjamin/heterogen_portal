package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderconfig"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/builderrepo"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/modelwriter"
	"github.com/turnerbenjamin/heterogen_portal/cmd/modelbuilder/schemamodel"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	config, err := builderconfig.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	cachePath, err := filepath.Abs("cmd/modelbuilder/metadata_cache.json")
	if err != nil {
		log.Fatal(err)
	}

	var metadata *builderrepo.DatabaseSchema = nil
	if config.DoUseCache {
		metadata, err = readFromCache(cachePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if metadata == nil {
		dbConnection := initDBConnection(config.DSN)
		defer closeDBConnection(dbConnection)

		metadataRepo := builderrepo.MetadataRepo{Db: dbConnection}
		metadata, err = metadataRepo.GetModelMetadata(
			config.BuilderConfig.ModelSchemaName,
			config.BuilderConfig.EnumSchemaName,
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

	enumWriter := modelwriter.NewEnumWriter(
		config.Paths.Enums,
		schemaData.EnumData,
	)
	enumWriter.Write()

	constantsWriter := modelwriter.NewConstantsWriter(
		config.Paths.Constants,
		schemaData,
	)
	constantsWriter.Write()

	queryModelWriter := modelwriter.NewQueryModelWriter(
		config.Paths.QueryModel,
		schemaData,
	)
	if err := queryModelWriter.Write(); err != nil {
		log.Fatal(err)
	}

	tableColumnDefinitionWriter := modelwriter.NewTableDefinitionWriter(
		config.BuilderConfig,
		config.Paths.TableColumnDefinitions,
		schemaData,
	)
	if err := tableColumnDefinitionWriter.Write(); err != nil {
		log.Fatal(err)
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
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(metadata)
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
