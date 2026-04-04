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

const (
	InputDir      = "data/input_pdfs"
	ProcessedFile = "data/processed.txt"
	ExtractDir    = "data/extracted_text"
	NormalizeDir  = "data/normalized_text"
	ChunkDir      = "data/chunks"
)

func setupEnvironment() {
	dirs := []string{InputDir, ExtractDir, NormalizeDir, ChunkDir}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
}

func saveJSON(v interface{}, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

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
	fmt.Println("🚀 READFLOW: Structural Intelligence Engine Started")

	// Ensure all directories exist so the "One Command" never fails
	setupEnvironment()

	processed := loadProcessed(ProcessedFile)

	files, err := os.ReadDir(InputDir)
	if err != nil {
		fmt.Printf("❌ Critical Error: %v\n", err)
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".pdf") || processed[f.Name()] {
			continue
		}

		fmt.Printf("\n📄 Processing: %s\n", f.Name())

		// --- STEP 1: SPATIAL EXTRACTION ---
		fmt.Print("  [1/3] Mapping document structure (X/Y Analysis)... ")
		doc, err := extract.ExtractText(filepath.Join(InputDir, f.Name()))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		saveJSON(doc, filepath.Join(ExtractDir, strings.TrimSuffix(f.Name(), ".pdf")+".json"))
		fmt.Println("Done.")

		// --- STEP 2: NORMALIZATION ---
		fmt.Print("  [2/3] Refining text & cleaning artifacts... ")
		rawPath := filepath.Join(ExtractDir, strings.TrimSuffix(f.Name(), ".pdf")+".json")
		err = normalize.NormalizeDocument(rawPath, NormalizeDir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Println("Done.")

		// --- STEP 3: STRUCTURAL CHUNKING ---
		fmt.Print("  [3/3] Gating signal & generating audit-ready chunks... ")
		normalizedPath := filepath.Join(NormalizeDir, strings.TrimSuffix(f.Name(), ".pdf")+".json")
		normalizedDoc, err := loadDocument(normalizedPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		err = chunk.ChunkDocument(*normalizedDoc, ChunkDir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Println("Done.")

		// Finalize
		markProcessed(ProcessedFile, f.Name())
		processed[f.Name()] = true
		fmt.Printf("✅ Success: %s is now refined and ready for inference.\n", f.Name())
	}
}
