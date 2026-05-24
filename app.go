package main

import (
	"context"
	"cria/internal/llm"
	"cria/internal/logging"
	"cria/internal/ollama"
	"cria/internal/pipeline"
	"cria/internal/vcs"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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
	logging.Infof("Starting Cria agent...")

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
			logging.Errorf("Fatal error: %v", err)
			return
		}
		a.serverCmd = cmd
	}()
}

func (a *App) shutdown(ctx context.Context) {
	logging.Infof("Shutting down Cria...")
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
	logging.Infof("[APP] StartUpgradePipeline initiated with task: %s", task)
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	logging.Infof("[APP] Workspace path: %s", workspacePath)

	err = vcs.SetupShadowWorkspace(cwd, workspacePath)
	if err != nil {
		logging.Errorf("[APP] SetupShadowWorkspace failed: %v", err)
		return
	}
	logging.Infof("[APP] Clone successful, starting orchestrator")

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
	logging.Infof("[APP] Rollback requested for hash: %s", hash)
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	err := gitMgr.RollbackToHash(hash)
	if err != nil {
		logging.Errorf("[APP] Rollback failed: %v", err)
		return err
	}

	return nil
}

func (a *App) GetUpgradeHistory() []vcs.UpgradeHistory {
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	history, err := gitMgr.GetUpgradeHistory()
	if err != nil {
		logging.Errorf("[APP] Error fetching history: %v", err)
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
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	if strings.HasSuffix(strings.ToLower(execPath), "-dev.exe") || CurrentCommit == "dev-mode-hash" {
		a.SimulateApplyAndRestart(hash, version)
		return nil
	}

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	checkoutCmd := exec.Command("git", "checkout", hash)
	checkoutCmd.Dir = workspacePath
	checkoutCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("checkout failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err == nil {
		fetchCmd := exec.Command("git", "fetch", workspacePath, hash)
		fetchCmd.Dir = cwd
		fetchCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = fetchCmd.Run()

		resetCmd := exec.Command("git", "reset", "--hard", hash)
		resetCmd.Dir = cwd
		resetCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = resetCmd.Run()

		tagsCmd := exec.Command("git", "fetch", workspacePath, "--tags")
		tagsCmd.Dir = cwd
		tagsCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = tagsCmd.Run()
	}

	ldflags := fmt.Sprintf("-X main.CurrentCommit=%s -X main.CurrentVersion=%s", hash, version)

	buildCmd := exec.Command("wails", "build", "-clean", "-ldflags", ldflags, "-o", "cria-upgrade.exe")
	buildCmd.Dir = workspacePath
	buildCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if out, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wails build failed: %v, output: %s", err, string(out))
	}

	newBinPath := filepath.Join(workspacePath, "build", "bin", "cria-upgrade.exe")
	execDir := filepath.Dir(execPath)

	oldExecPath := execPath + ".old"
	_ = os.Remove(oldExecPath)
	if err := os.Rename(execPath, oldExecPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %v", err)
	}

	moveCmd := exec.Command("cmd", "/c", "move", "/Y", newBinPath, execPath)
	moveCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := moveCmd.Run(); err != nil {
		_ = os.Rename(oldExecPath, execPath)
		return fmt.Errorf("failed to move new binary: %v", err)
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

func (a *App) SimulateApplyAndRestart(hash string, version string) {
	logging.Infof("[APP] Simulating restart. Updating CurrentCommit to: %s, Version: %s", hash, version)
	CurrentCommit = hash
	CurrentVersion = version
}
