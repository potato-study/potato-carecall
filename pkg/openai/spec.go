package openai

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	functionSpecs = loadFunctions()
	promptFile    = "config/prompt"
	functionPath  = "config/function/"
)

func PromptMessage() Message {
	content, err := os.ReadFile(promptFile)
	if err != nil {
		log.Fatalf("Failed to read prompt file: %v", err)
	}
	return Message{
		Role:    "system",
		Content: string(content),
	}
}

func loadFunctions() []FunctionSpec {
	var specs []FunctionSpec

	files, err := os.ReadDir(functionPath)
	if err != nil {
		log.Fatalf("Failed to read function directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(functionPath, file.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file %s: %v", filePath, err)
			continue
		}

		var commonParams map[string]interface{}
		if err := json.Unmarshal(fileContent, &commonParams); err != nil {
			log.Printf("Failed to unmarshal JSON from file %s: %v", filePath, err)
			continue
		}

		specs = append(specs, FunctionSpec{
			Name:        strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			Description: "Description for " + file.Name(),
			Parameters:  commonParams,
		})
	}

	return specs
}
