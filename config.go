package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	OllamaModelsPath string `json:"ollama_models_path"`
}

func loadConfigPath() string {
	file, err := os.Open("config.json")
	if err != nil {
		return ""
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return ""
	}
	return cfg.OllamaModelsPath
}

func saveConfigPath(path string) {
	cfg := Config{OllamaModelsPath: path}
	file, err := os.Create("config.json")
	if err != nil {
		return
	}
	defer file.Close()

	_ = json.NewEncoder(file).Encode(cfg)
}
