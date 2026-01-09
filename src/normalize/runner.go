package normalize

import (
	"encoding/json"
	"os"
	"path/filepath"

	"readflow/src/extract"
)

func NormalizeDocument(inputPath string, outputDir string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var doc extract.DocumentText
	if err := json.NewDecoder(file).Decode(&doc); err != nil {
		return err
	}

	for i := range doc.Pages {
		doc.Pages[i].Text = NormalizeText(doc.Pages[i].Text)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	outPath := filepath.Join(outputDir, filepath.Base(inputPath))

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")

	return encoder.Encode(doc)
}
