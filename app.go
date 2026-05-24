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

var InitialCommit string = ""
var CurrentCommit string = ""
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

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	cwdHead := "dev-mode-hash"
	if err == nil {
		cwdHead = strings.TrimSpace(string(out))
	}

	if InitialCommit == "" {
		InitialCommit = cwdHead
	}
	if CurrentCommit == "" {
		CurrentCommit = cwdHead
	}

	logging.Userf("app.startup InitialCommit=%s CurrentCommit=%s CurrentVersion=%s", InitialCommit, CurrentCommit, CurrentVersion)

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	if _, err := os.Stat(filepath.Join(workspacePath, ".git")); os.IsNotExist(err) {
		logging.Statef("workspace not found. creating shadow workspace at %s", workspacePath)
		_ = vcs.SetupShadowWorkspace(cwd, workspacePath)
	}

	path := loadConfigPath()
	if path != "" {
		os.Setenv("OLLAMA_MODELS", path)
		logging.Statef("OLLAMA_MODELS set from config: %s", path)
	} else if os.Getenv("OLLAMA_MODELS") == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			globalModelsPath := filepath.Join(homeDir, ".ollama", "models")
			os.Setenv("OLLAMA_MODELS", globalModelsPath)
			saveConfigPath(globalModelsPath)
			logging.Statef("OLLAMA_MODELS defaulted: %s", globalModelsPath)
		}
	}

	go func() {
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			logging.Errorf("ollama EnsureInstalledAndRun: %v", err)
			return
		}
		a.serverCmd = cmd
		logging.Statef("ollama runner started pid=%d", cmd.Process.Pid)
	}()
}

func (a *App) shutdown(ctx context.Context) {
	logging.Userf("app.shutdown")
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
		logging.Statef("ollama runner killed")
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetOllamaPath() string {
	return os.Getenv("OLLAMA_MODELS")
}

func (a *App) UpdateOllamaPath(newPath string) bool {
	logging.Userf("UpdateOllamaPath newPath=%s", newPath)
	os.Setenv("OLLAMA_MODELS", newPath)
	saveConfigPath(newPath)

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
		logging.Statef("ollama runner killed for restart")
	}

	go func() {
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			logging.Errorf("ollama restart failed: %v", err)
			return
		}
		a.serverCmd = cmd
		logging.Statef("ollama runner restarted pid=%d", cmd.Process.Pid)
	}()

	return true
}

func (a *App) SelectFolder() string {
	logging.Userf("SelectFolder dialog opened")
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Ollama Models Directory",
	})
	if err != nil {
		logging.Errorf("SelectFolder dialog: %v", err)
		return ""
	}
	logging.Userf("SelectFolder picked: %s", folder)
	return folder
}

func (a *App) GetOllamaModels() []string {
	models := llm.FetchOllamaModels()
	logging.Debugf("GetOllamaModels -> %d models", len(models))
	return models
}

func (a *App) DownloadModel(modelName string) string {
	logging.Userf("DownloadModel model=%s", modelName)
	return llm.DownloadOllamaModel(a.ctx, modelName)
}

func (a *App) ChatWithModel(modelName string, prompt string) string {
	logging.Userf("ChatWithModel model=%s promptLen=%d", modelName, len(prompt))
	return llm.ChatWithOllama(modelName, prompt)
}

func (a *App) RemoveModel(modelName string) string {
	logging.Userf("RemoveModel model=%s", modelName)
	return llm.RemoveOllamaModel(modelName)
}

func (a *App) StartUpgradePipeline(task string) {
	logging.Userf("StartUpgradePipeline task=%q", task)
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
		logging.Errorf("os.Getwd: %v (using .)", err)
	}

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	logging.Statef("workspace path: %s (cwd=%s)", workspacePath, cwd)

	err = vcs.SetupShadowWorkspace(cwd, workspacePath)
	if err != nil {
		logging.Errorf("SetupShadowWorkspace: %v", err)
		return
	}
	logging.Statef("workspace ready, launching orchestrator")

	orc := pipeline.NewOrchestrator(a.ctx, workspacePath)
	go orc.RunMock(task, a.hitlChan)
}

func (a *App) ApproveHITL() {
	logging.Userf("ApproveHITL")
	a.hitlChan <- pipeline.HITLResponse{Approved: true}
}

func (a *App) RejectHITL(feedback string) {
	logging.Userf("RejectHITL feedback=%q", feedback)
	a.hitlChan <- pipeline.HITLResponse{Approved: false, Feedback: feedback}
}

func (a *App) RollbackUpgrade(hash string) error {
	logging.Userf("RollbackUpgrade hash=%s", hash)
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	err := gitMgr.RollbackToHash(hash)
	if err != nil {
		logging.Errorf("RollbackToHash workspace: %v", err)
		return err
	}

	logging.Statef("rollback completed for hash=%s", hash)
	return nil
}

func (a *App) GetUpgradeHistory() []vcs.UpgradeHistory {
	logging.Userf("GetUpgradeHistory")
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	gitMgr := vcs.NewGitManager(workspacePath)
	history, err := gitMgr.GetUpgradeHistory()
	if err != nil {
		logging.Errorf("GetUpgradeHistory: %v", err)
		return []vcs.UpgradeHistory{}
	}
	logging.Statef("history returned %d entries", len(history))
	for i, h := range history {
		logging.Debugf("  history[%d] hash=%s version=%s msg=%q", i, h.Hash, h.Version, h.Message)
	}
	return history
}

func (a *App) GetInitialCommit() string {
	logging.Debugf("GetInitialCommit -> %s", InitialCommit)
	return InitialCommit
}

func (a *App) GetActiveCommit() string {
	logging.Debugf("GetActiveCommit -> %s", CurrentCommit)
	return CurrentCommit
}

func (a *App) GetActiveVersion() string {
	logging.Debugf("GetActiveVersion -> %s", CurrentVersion)
	return CurrentVersion
}

func (a *App) ApplyUpgrade(hash string, version string) error {
	logging.Userf("ApplyUpgrade hash=%s version=%s", hash, version)

	execPath, err := os.Executable()
	if err != nil {
		logging.Errorf("os.Executable: %v", err)
		return err
	}
	logging.Statef("execPath=%s", execPath)

	if strings.HasSuffix(strings.ToLower(execPath), "-dev.exe") {
		logging.Statef("ApplyUpgrade dev mode path: simulating restart")
		a.SimulateApplyAndRestart(hash, version)
		return nil
	}

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	logging.Statef("ApplyUpgrade: checkout %s in %s", hash, workspacePath)
	checkoutCmd := exec.Command("git", "checkout", hash)
	checkoutCmd.Dir = workspacePath
	checkoutCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := checkoutCmd.Run(); err != nil {
		logging.Errorf("ApplyUpgrade checkout failed: %v", err)
		return fmt.Errorf("checkout failed: %v", err)
	}

	ldflags := fmt.Sprintf("-X main.InitialCommit=%s -X main.CurrentCommit=%s -X main.CurrentVersion=%s", InitialCommit, hash, version)
	logging.Statef("ApplyUpgrade running wails build with ldflags=%s", ldflags)

	buildCmd := exec.Command("wails", "build", "-clean", "-ldflags", ldflags, "-o", "cria-upgrade.exe")
	buildCmd.Dir = workspacePath
	buildCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if out, err := buildCmd.CombinedOutput(); err != nil {
		logging.Errorf("wails build failed: %v output=%s", err, string(out))
		return fmt.Errorf("wails build failed: %v, output: %s", err, string(out))
	}
	logging.Statef("wails build completed")

	newBinPath := filepath.Join(workspacePath, "build", "bin", "cria-upgrade.exe")
	execDir := filepath.Dir(execPath)

	oldExecPath := execPath + ".old"
	_ = os.Remove(oldExecPath)
	if err := os.Rename(execPath, oldExecPath); err != nil {
		logging.Errorf("backup current binary: %v", err)
		return fmt.Errorf("failed to backup current binary: %v", err)
	}
	logging.Statef("current binary backed up to %s", oldExecPath)

	moveCmd := exec.Command("cmd", "/c", "move", "/Y", newBinPath, execPath)
	moveCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := moveCmd.Run(); err != nil {
		_ = os.Rename(oldExecPath, execPath)
		logging.Errorf("move new binary into place: %v (reverted)", err)
		return fmt.Errorf("failed to move new binary: %v", err)
	}
	logging.Statef("new binary moved to %s", execPath)

	newCmd := exec.Command(execPath)
	newCmd.Dir = execDir
	if err := newCmd.Start(); err != nil {
		_ = os.Rename(oldExecPath, execPath)
		logging.Errorf("restart new binary: %v (reverted)", err)
		return fmt.Errorf("failed to restart application: %v", err)
	}
	logging.Statef("new binary started pid=%d, exiting current", newCmd.Process.Pid)

	go func() {
		os.Exit(0)
	}()

	return nil
}

func (a *App) GetLatestVersion() string {
	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	gitMgr := vcs.NewGitManager(workspacePath)
	v := gitMgr.GetLatestTag()
	logging.Debugf("GetLatestVersion -> %s", v)
	return v
}

func (a *App) SimulateApplyAndRestart(hash string, version string) {
	logging.Statef("SimulateApplyAndRestart hash=%s version=%s", hash, version)
	CurrentCommit = hash
	CurrentVersion = version
}
