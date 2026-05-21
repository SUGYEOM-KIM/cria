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

var CurrentCommit string = "dev-mode-hash"
var CurrentVersion string = "v0.0.0"

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
	fmt.Println("[APP] StartUpgradePipeline initiated with task:", task)
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	fmt.Println("[APP] Workspace path:", workspacePath)

	err = vcs.SetupShadowWorkspace(cwd, workspacePath)
	if err != nil {
		fmt.Printf("[APP] Error: SetupShadowWorkspace failed: %v\n", err)
		return
	}
	fmt.Println("[APP] Clone successful, starting orchestrator")

	orc := pipeline.NewOrchestrator(a.ctx, workspacePath)
	go orc.RunMock(task, a.hitlChan)
}

func (a *App) ApproveHITL() {
	a.hitlChan <- pipeline.HITLResponse{Approved: true}
}

func (a *App) RejectHITL(feedback string) {
	a.hitlChan <- pipeline.HITLResponse{Approved: false, Feedback: feedback}
}

func (a *App) RollbackUpgrade(hash string) error {
	fmt.Println("[APP] Rollback requested for hash:", hash)
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	err := gitMgr.RollbackToHash(hash)
	if err != nil {
		fmt.Printf("[APP] Rollback failed: %v\n", err)
		return err
	}

	return nil
}

func (a *App) GetUpgradeHistory() []vcs.UpgradeHistory {
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	history, err := gitMgr.GetUpgradeHistory()
	if err != nil {
		fmt.Printf("[APP] Error fetching history: %v\n", err)
		return []vcs.UpgradeHistory{}
	}
	return history
}

func (a *App) GetActiveCommit() string {
	return CurrentCommit
}

func (a *App) GetActiveVersion() string {
	return CurrentVersion
}

func (a *App) ApplyUpgrade(hash string, version string) error {
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	checkoutCmd := exec.Command("git", "checkout", hash)
	checkoutCmd.Dir = workspacePath
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("checkout failed: %v", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	oldExecPath := execPath + ".old"
	_ = os.Remove(oldExecPath)
	if err := os.Rename(execPath, oldExecPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %v", err)
	}

	ldflags := fmt.Sprintf("-X main.CurrentCommit=%s -X main.CurrentVersion=%s", hash, version)
	buildCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", execPath, ".")
	buildCmd.Dir = workspacePath
	if out, err := buildCmd.CombinedOutput(); err != nil {
		_ = os.Rename(oldExecPath, execPath)
		return fmt.Errorf("build failed: %v, output: %s", err, string(out))
	}

	newCmd := exec.Command(execPath)
	newCmd.Dir = execDir
	if err := newCmd.Start(); err != nil {
		_ = os.Rename(oldExecPath, execPath)
		return fmt.Errorf("failed to restart application: %v", err)
	}

	go func() {
		os.Exit(0)
	}()

	return nil
}

func (a *App) GetLatestVersion() string {
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	gitMgr := vcs.NewGitManager(workspacePath)
	return gitMgr.GetLatestTag()
}
