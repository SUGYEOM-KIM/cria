package main

import (
	"context"
	"cria/internal/llm"
	"cria/internal/ollama"
	"cria/internal/pipeline"
	"cria/internal/vcs"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	serverCmd *exec.Cmd
	hitlChan  chan pipeline.HITLResponse
}

func NewApp() *App {
	return &App{
		hitlChan: make(chan pipeline.HITLResponse),
	}
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
	saveConfigPath(newPath)

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
	}

	go func() {
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
	return llm.FetchOllamaModels()
}

func (a *App) DownloadModel(modelName string) string {
	return llm.DownloadOllamaModel(a.ctx, modelName)
}

func (a *App) ChatWithModel(modelName string, prompt string) string {
	return llm.ChatWithOllama(modelName, prompt)
}

func (a *App) RemoveModel(modelName string) string {
	return llm.RemoveOllamaModel(modelName)
}

func (a *App) StartUpgradePipeline(task string) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	workspacePath := filepath.Join(cwd, ".cria_workspace")

	err = vcs.SetupShadowWorkspace(cwd, workspacePath)
	if err != nil {
		fmt.Printf("Failed to setup shadow workspace: %v\n", err)
		return
	}

	orc := pipeline.NewOrchestrator(a.ctx, workspacePath)
	go orc.RunMock(task, a.hitlChan)
}

func (a *App) ApproveHITL() {
	a.hitlChan <- pipeline.HITLResponse{Approved: true}
}

func (a *App) RejectHITL(feedback string) {
	a.hitlChan <- pipeline.HITLResponse{Approved: false, Feedback: feedback}
}
