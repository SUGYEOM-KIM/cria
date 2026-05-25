package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	OllamaModelsPath string            `json:"ollama_models_path"`
	AgentModels      map[string]string `json:"agent_models,omitempty"`
}

func loadConfig() Config {
	var cfg Config
	file, err := os.Open("config.json")
	if err != nil {
		return cfg
	}
	defer file.Close()
	_ = json.NewDecoder(file).Decode(&cfg)
	return cfg
}

func saveConfig(cfg Config) {
	file, err := os.Create("config.json")
	if err != nil {
		return
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cfg)
}

func loadConfigPath() string {
	return loadConfig().OllamaModelsPath
}

func saveConfigPath(path string) {
	cfg := loadConfig()
	cfg.OllamaModelsPath = path
	saveConfig(cfg)
}

func loadAgentModels() map[string]string {
	models := loadConfig().AgentModels
	if models == nil {
		return map[string]string{}
	}
	return models
}

func saveAgentModels(models map[string]string) {
	cfg := loadConfig()
	cfg.AgentModels = models
	saveConfig(cfg)
}
