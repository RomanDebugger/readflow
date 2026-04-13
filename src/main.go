package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func processHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename
	fmt.Printf("\nIncoming HTTP Request: Processing %s\n", filename)

	inputPath := filepath.Join(InputDir, filename)
	outFile, err := os.Create(inputPath)
	if err != nil {
		http.Error(w, "Failed to save temp file", http.StatusInternalServerError)
		return
	}
	io.Copy(outFile, file)
	outFile.Close()

	fmt.Print("  [1/3] Mapping document structure... ")
	doc, err := extract.ExtractText(inputPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Extraction error: %v", err), http.StatusInternalServerError)
		return
	}
	rawPath := filepath.Join(ExtractDir, strings.TrimSuffix(filename, ".pdf")+".json")
	saveJSON(doc, rawPath)
	fmt.Println("Done.")

	fmt.Print("  [2/3] Refining text... ")
	err = normalize.NormalizeDocument(rawPath, NormalizeDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Normalization error: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("Done.")

	fmt.Print("  [3/3] Generating chunks... ")
	normalizedPath := filepath.Join(NormalizeDir, strings.TrimSuffix(filename, ".pdf")+".json")
	normalizedDoc, err := loadDocument(normalizedPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Loading normalized doc error: %v", err), http.StatusInternalServerError)
		return
	}

	err = chunk.ChunkDocument(*normalizedDoc, ChunkDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Chunking error: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("Done.")

	finalChunkPath := filepath.Join(ChunkDir, strings.TrimSuffix(filename, ".pdf")+".json")
	finalData, err := os.ReadFile(finalChunkPath)
	if err != nil {
		http.Error(w, "Failed to read final chunks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(finalData)
	fmt.Printf("Success: %s processed and sent back to frontend.\n", filename)
}

func main() {
	setupEnvironment()

	http.HandleFunc("/process", processHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("READFLOW: Go Engine running and listening on port %s...\n", port)

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("Critical Server Error: %v\n", err)
	}
}
