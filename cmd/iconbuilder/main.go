package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/turnerbenjamin/heterogen_portal/cmd/iconbuilder/iconbuilderutils"
)

type config struct {
	sourceDir  string
	targetPath string
}

func main() {

	config := readConfig()
	svgData, err := iconbuilderutils.ReadIcons(config.sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	iconbuilderutils.WriteIcons(config.targetPath, svgData)

	// step through files

	// write to the builder
}

func readConfig() config {
	sourceDir := flag.String("source-dir", "", "source-dir")
	targetPath := flag.String("target-path", "", "target path")

	flag.Parse()

	if *sourceDir == "" {
		log.Fatal("source dir must be set")
	}

	if *targetPath == "" {
		log.Fatal("target path must be set")
	}

	absSourceDir, err := filepath.Abs(*sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	absTargetPath, err := filepath.Abs(*targetPath)
	if err != nil {
		log.Fatal(err)
	}

	return config{
		sourceDir:  absSourceDir,
		targetPath: absTargetPath,
	}
}
