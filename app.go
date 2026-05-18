package main

import (
	"bufio"
	"context"
	"cria/internal/ollama"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Config struct {
	OllamaModelsPath string `json:"ollama_models_path"`
}

type App struct {
	ctx       context.Context
	serverCmd *exec.Cmd
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Starting Cria agent...")

	path := a.loadConfigPath()
	if path != "" {
		os.Setenv("OLLAMA_MODELS", path)
	} else if os.Getenv("OLLAMA_MODELS") == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			globalModelsPath := filepath.Join(homeDir, ".ollama", "models")
			os.Setenv("OLLAMA_MODELS", globalModelsPath)
			a.saveConfigPath(globalModelsPath)
		}
	}

	fmt.Printf("Current OLLAMA_MODELS path: %s\n", os.Getenv("OLLAMA_MODELS"))

	go func() {
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			fmt.Printf("Fatal error: %v\n", err)
			return
		}
		a.serverCmd = cmd
	}()
}

func (a *App) shutdown(ctx context.Context) {
	fmt.Println("Shutting down Cria...")
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetOllamaPath() string {
	return os.Getenv("OLLAMA_MODELS")
}

func (a *App) UpdateOllamaPath(newPath string) bool {
	os.Setenv("OLLAMA_MODELS", newPath)
	a.saveConfigPath(newPath)

	fmt.Printf("OLLAMA_MODELS Env Set To: %s\n", os.Getenv("OLLAMA_MODELS"))

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		fmt.Printf("Killing existing Ollama process (PID: %d)...\n", a.serverCmd.Process.Pid)
		_ = a.serverCmd.Process.Kill()
	} else {
		fmt.Printf("No existing Ollama process found to kill.\n")
	}

	go func() {
		fmt.Printf("Restarting Ollama engine...\n")
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			fmt.Printf("Fatal error restarting Ollama: %v\n", err)
			return
		}
		a.serverCmd = cmd
		fmt.Printf("Ollama engine restarted successfully! (PID: %d)\n", cmd.Process.Pid)
	}()

	return true
}

type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (a *App) GetOllamaModels() []string {
	fmt.Println("[DEBUG] Fetching models from Ollama API...")
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Printf("[DEBUG] HTTP request error: %v\n", err)
		return []string{}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[DEBUG] Read body error: %v\n", err)
		return []string{}
	}

	fmt.Printf("[DEBUG] Raw response from Ollama: %s\n", string(bodyBytes))

	var tagsResp OllamaTagsResponse
	if err := json.Unmarshal(bodyBytes, &tagsResp); err != nil {
		fmt.Printf("[DEBUG] JSON decode error: %v\n", err)
		return []string{}
	}

	var models []string
	for _, m := range tagsResp.Models {
		models = append(models, m.Name)
	}

	fmt.Printf("[DEBUG] Successfully fetched %d models.\n", len(models))
	return models
}

func (a *App) loadConfigPath() string {
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

func (a *App) saveConfigPath(path string) {
	cfg := Config{OllamaModelsPath: path}
	file, err := os.Create("config.json")
	if err != nil {
		return
	}
	defer file.Close()

	_ = json.NewEncoder(file).Encode(cfg)
}

func (a *App) SelectFolder() string {
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Ollama Models Directory",
	})
	if err != nil {
		return ""
	}
	return folder
}

func (a *App) DownloadModel(modelName string) string {
	cmd := exec.Command("ollama", "pull", modelName)

	if currentPath := os.Getenv("OLLAMA_MODELS"); currentPath != "" {
		cmd.Env = append(os.Environ(), "OLLAMA_MODELS="+currentPath)
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("Error starting: %v", err)
	}

	go a.processOutput(stdout, modelName)
	go a.processOutput(stderr, modelName)

	err = cmd.Wait()
	if err != nil {
		return fmt.Sprintf("Error finishing: %v", err)
	}

	runtime.EventsEmit(a.ctx, "download-progress-"+modelName, "100%")
	return "Success"
}

func (a *App) processOutput(pipe io.ReadCloser, modelName string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "%") {
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.Contains(part, "%") {
					runtime.EventsEmit(a.ctx, "download-progress-"+modelName, part)
					break
				}
			}
		} else {
			runtime.EventsEmit(a.ctx, "download-progress-"+modelName, line)
		}
	}
}
