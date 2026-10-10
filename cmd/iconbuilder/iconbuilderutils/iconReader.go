package iconbuilderutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type svgData struct {
	name     string
	contents string
}

func ReadIcons(sourceDir string) (svgs []svgData, err error) {
	svgs = []svgData{}

	dir, err := os.Open(sourceDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = dir.Close()
	}()

	files, err := dir.ReadDir(0)
	if err != nil {
		log.Fatal(err)
	}

	for _, dirEntry := range files {
		if dirEntry.IsDir() {
			continue
		}

		if !strings.EqualFold(filepath.Ext(dirEntry.Name()), ".svg") {
			continue
		}

		filePath := filepath.Join(sourceDir, dirEntry.Name())

		contents, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		svgName := strings.Replace(dirEntry.Name(), filepath.Ext(dirEntry.Name()), "", 1)
		svgPascalName, err := fnToPascal(svgName)
		if err != nil {
			fmt.Printf("unable to parse svg: %s", err.Error())
			continue
		}

		svgs = append(svgs, svgData{
			name:     svgPascalName,
			contents: string(contents),
		})
	}
	return svgs, nil
}

func fnToPascal(s string) (string, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for i, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) == 0 {
			continue
		}

		if i == 0 && unicode.IsDigit(runes[0]) {
			return "", fmt.Errorf("names cannot start with a digit: %s", s)
		}

		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	return strings.Join(parts, ""), nil
}
