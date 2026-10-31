package builderconfig

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path"
	"path/filepath"
)

type TableColumnDefinitionConfig struct {
	TableName           string   `json:"table-name"`
	Version             uint64   `json:"version"`
	ExcludeFromMainGrid []string `json:"exclude-from-main-grid"`
	RelatedColumns      []string `json:"related-columns"`
}

type BuilderConfig struct {
	ModelSchemaName  string                                 `json:"model-schema-name"`
	EnumSchemaName   string                                 `json:"enum-schema-name"`
	TableDefinitions map[string]TableColumnDefinitionConfig `json:"table-column-definitions"`
}

type Paths struct {
	QueryModel             string
	Enums                  string
	Constants              string
	TableColumnDefinitions string
}

type Config struct {
	DSN           string
	ModelDir      string
	DoUseCache    bool
	BuilderConfig BuilderConfig
	Paths         Paths
}

func GetConfig() (Config, error) {
	serverDsn := flag.String("dsn", "", "server dsn")
	modelDir := flag.String("model-dir", "", "directory to write model to")
	builderConfigPath := flag.String("builder-config", "", "path to builder config")
	useCache := flag.Bool("use-cache", false, "use cached model data")

	flag.Parse()

	if *serverDsn == "" {
		return Config{}, errors.New("invalid command: -dsn flag must be set")
	}

	if *modelDir == "" {
		return Config{}, errors.New("invalid command: -model-dir flag must be set")
	}

	if *builderConfigPath == "" {
		return Config{}, errors.New("invalid command: -builder-config flag must be set")
	}

	paths, err := getPaths(*modelDir)
	if err != nil {
		return Config{}, err
	}

	builderConfig, err := parseBuilderConfig(*builderConfigPath)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DSN:           *serverDsn,
		ModelDir:      *modelDir,
		BuilderConfig: builderConfig,
		DoUseCache:    *useCache,
		Paths:         paths,
	}, nil
}

func parseBuilderConfig(builderConfig string) (config BuilderConfig, err error) {
	p, err := filepath.Abs(builderConfig)
	if err != nil {
		return BuilderConfig{}, err
	}

	configData, err := os.ReadFile(p)
	if err != nil {
		return BuilderConfig{}, err
	}

	var out BuilderConfig
	err = json.Unmarshal(configData, &out)
	if err != nil {
		return BuilderConfig{}, err
	}

	if out.ModelSchemaName == "" {
		return out, errors.New("invalid config: model-schema-name is required")
	}

	if out.EnumSchemaName == "" {
		return out, errors.New("invalid config: enum-schema-name is required")
	}

	return out, nil
}

func getPaths(modelDir string) (Paths, error) {
	d, err := filepath.Abs(modelDir)
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		QueryModel:             path.Join(d, "queryModel.go"),
		Enums:                  path.Join(d, "enums.go"),
		Constants:              path.Join(d, "constants.go"),
		TableColumnDefinitions: path.Join(d, "tableDefinitions.go"),
	}, nil
}
