package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"readflow/src/extract"
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

		err = markProcessed(processedFile, f.Name())
		if err != nil {
			fmt.Println("Failed to mark processed:", err)
			continue
		}

		processed[f.Name()] = true
	}
}
