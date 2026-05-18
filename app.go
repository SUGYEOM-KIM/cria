package main

import (
	"context"
	"cria/internal/ollama"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

	path := loadConfigPath()
	if path != "" {
		os.Setenv("OLLAMA_MODELS", path)
	} else if os.Getenv("OLLAMA_MODELS") == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			globalModelsPath := filepath.Join(homeDir, ".ollama", "models")
			os.Setenv("OLLAMA_MODELS", globalModelsPath)
			saveConfigPath(globalModelsPath)
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

// ----------------------------------------------------
// Frontend Exposed Functions (Wails Bindings)
// ----------------------------------------------------

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetOllamaPath() string {
	return os.Getenv("OLLAMA_MODELS")
}

func (a *App) UpdateOllamaPath(newPath string) bool {
	os.Setenv("OLLAMA_MODELS", newPath)
	saveConfigPath(newPath)

	fmt.Printf("OLLAMA_MODELS Env Set To: %s\n", os.Getenv("OLLAMA_MODELS"))

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		fmt.Printf("Killing existing Ollama process (PID: %d)...\n", a.serverCmd.Process.Pid)
		_ = a.serverCmd.Process.Kill()
	}

	go func() {
		fmt.Printf("Restarting Ollama engine...\n")
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			return
		}
		a.serverCmd = cmd
	}()

	return true
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

func (a *App) GetOllamaModels() []string {
	return fetchOllamaModels()
}

func (a *App) DownloadModel(modelName string) string {
	return downloadOllamaModel(a.ctx, modelName)
}

func (a *App) ChatWithModel(modelName string, prompt string) string {
	return chatWithOllama(modelName, prompt)
}

func (a *App) RemoveModel(modelName string) string {
	return removeOllamaModel(modelName)
}
