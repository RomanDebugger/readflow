package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"readflow/src/chunk"
	"readflow/src/extract"
	"readflow/src/normalize"
	"strings"
)

func loadProcessed(path string) map[string]bool {
	processed := make(map[string]bool)
	file, err := os.Open(path)

	if err != nil {
		return processed
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			processed[name] = true
		}
	}
	return processed
}

func markProcessed(path string, filename string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(filename + "\n")
	return err
}

func saveExtracted(doc *extract.DocumentText, outDir string) error {
	os.MkdirAll(outDir, 0755)

	base := filepath.Base(doc.Document)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	outPath := filepath.Join(outDir, name+".json")

	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(doc)
}

func loadDocument(path string) (*extract.DocumentText, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var doc extract.DocumentText
	err = json.NewDecoder(file).Decode(&doc)
	return &doc, err
}

func main() {
	fmt.Println("readflow started")

	inputDir := "data/input_pdfs"
	processedFile := "data/processed.txt"

	processed := loadProcessed(processedFile)

	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Println("Error reading input directory:", err)
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}

		if processed[f.Name()] {
			continue
		}

		fmt.Println("New PDF detected:", f.Name())
		doc, err := extract.ExtractText(inputDir + "/" + f.Name())
		if err != nil {
			fmt.Println("Error extracting text:", err)
			continue
		}

		err = saveExtracted(doc, "data/extracted_text")
		if err != nil {
			fmt.Println("Error saving extracted text:", err)
			continue
		}

		rawPath := filepath.Join("data/extracted_text",
			strings.TrimSuffix(f.Name(), ".pdf")+".json")

		err = normalize.NormalizeDocument(rawPath, "data/normalized_text")
		if err != nil {
			fmt.Println("Normalization failed:", err)
			continue
		}

		normalizedPath := filepath.Join(
			"data/normalized_text",
			strings.TrimSuffix(f.Name(), ".pdf")+".json",
		)

		normalizedDoc, err := loadDocument(normalizedPath)
		if err != nil {
			fmt.Println("Failed to load normalized doc:", err)
			continue
		}

		err = chunk.ChunkDocument(*normalizedDoc, "data/chunks")
		if err != nil {
			fmt.Println("Chunking failed:", err)
			continue
		}

		err = markProcessed(processedFile, f.Name())
		if err != nil {
			fmt.Println("Failed to mark processed:", err)
			continue
		}

		processed[f.Name()] = true
	}
}
