package main

import (
	"context"
	"cria/internal/agent"
	"cria/internal/llm"
	"cria/internal/logging"
	"cria/internal/ollama"
	"cria/internal/pipeline"
	"cria/internal/vcs"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed source.zip
var sourceZip []byte
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

	runtime.WindowShow(ctx)
	runtime.WindowUnminimise(ctx)
	runtime.WindowSetAlwaysOnTop(ctx, true)
	runtime.WindowSetAlwaysOnTop(ctx, false)

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")
	if _, err := os.Stat(filepath.Join(workspacePath, ".git")); os.IsNotExist(err) {
		logging.Statef("workspace not found. extracting embedded source to %s", workspacePath)
		_ = vcs.SetupWorkspaceFromZip(sourceZip, workspacePath)
	}

	if CurrentCommit == "" {
		workspaceHead := "dev-mode-hash"
		headCmd := exec.Command("git", "rev-parse", "HEAD")
		headCmd.Dir = workspacePath
		headCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if out, err := headCmd.Output(); err == nil {
			workspaceHead = strings.TrimSpace(string(out))
		}
		CurrentCommit = workspaceHead
	}

	logging.Userf("app.startup CurrentCommit=%s CurrentVersion=%s", CurrentCommit, CurrentVersion)

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

		if llm.WaitForReady(a.ctx, 30*time.Second) {
			logging.Statef("ollama ready, emitting ollama-ready")
			runtime.EventsEmit(a.ctx, "ollama-ready")
		} else {
			logging.Errorf("ollama did not become ready within timeout")
		}
	}()
}

func (a *App) shutdown(_ context.Context) {
	logging.Userf("app.shutdown")
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
		logging.Statef("ollama runner killed")
	}
}

func (a *App) LogClientEvent(level string, message string) {
	switch level {
	case "error":
		logging.Errorf("[CLIENT] %s", message)
	case "debug":
		logging.Debugf("[CLIENT] %s", message)
	case "state":
		logging.Statef("[CLIENT] %s", message)
	default:
		logging.Userf("[CLIENT] %s", message)
	}
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
	}

	_ = exec.Command("taskkill", "/F", "/IM", "ollama.exe").Run()
	_ = exec.Command("taskkill", "/F", "/IM", "ollama app.exe").Run()

	logging.Statef("force killed background ollama processes. waiting for port 11434 release...")
	time.Sleep(1 * time.Second)

	cmd, err := ollama.EnsureInstalledAndRun()
	if err != nil {
		logging.Errorf("ollama restart failed: %v", err)
		return false
	}
	a.serverCmd = cmd
	logging.Statef("ollama runner restarted pid=%d", cmd.Process.Pid)

	if llm.WaitForReady(a.ctx, 30*time.Second) {
		logging.Statef("ollama ready after restart, emitting ollama-ready")
		runtime.EventsEmit(a.ctx, "ollama-ready")
		return true
	} else {
		logging.Errorf("ollama did not become ready within timeout after restart")
		return false
	}
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
	if len(models) == 0 {
		if llm.WaitForReady(a.ctx, 2*time.Second) {
			models = llm.FetchOllamaModels()
		}
	}
	logging.Debugf("GetOllamaModels -> %d models", len(models))

	if models == nil {
		return []string{}
	}
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

func (a *App) GetAgentModels() map[string]string {
	models := loadAgentModels()
	logging.Debugf("GetAgentModels -> %d entries", len(models))
	return models
}

func (a *App) SaveAgentModels(models map[string]string) bool {
	logging.Userf("SaveAgentModels entries=%d", len(models))
	saveAgentModels(models)
	return true
}

func (a *App) GetTranslationLanguage() string {
	lang := loadTranslationLanguage()
	logging.Debugf("GetTranslationLanguage -> %q", lang)
	return lang
}

func (a *App) SaveTranslationLanguage(lang string) bool {
	logging.Userf("SaveTranslationLanguage lang=%q", lang)
	saveTranslationLanguage(lang)
	return true
}

func (a *App) TranslateText(model string, targetLang string, text string) string {
	logging.Userf("TranslateText model=%s lang=%s textLen=%d", model, targetLang, len(text))

	resolvedModel := model
	if resolvedModel == "" {
		agentModels := loadAgentModels()
		if v, ok := agentModels["translator"]; ok && v != "" {
			resolvedModel = v
		} else if v, ok := agentModels["global"]; ok && v != "" {
			resolvedModel = v
		}
	}
	if resolvedModel == "" {
		logging.Errorf("TranslateText: no model resolved")
		return "Error: no translation model configured."
	}

	resolvedLang := targetLang
	if resolvedLang == "" {
		resolvedLang = loadTranslationLanguage()
	}
	if resolvedLang == "" {
		logging.Errorf("TranslateText: no target language configured")
		return "Error: no target language configured. Please set one in Settings."
	}

	systemPrompt := "You are a professional translator. Translate the user's text into " + resolvedLang + ". Preserve all Markdown formatting, code blocks, and structure exactly. Do not add explanations, notes, or commentary. Respond with only the translated text."
	fullPrompt := systemPrompt + "\n\n" + text

	logging.Statef("TranslateText calling model=%s lang=%s", resolvedModel, resolvedLang)
	return llm.ChatWithOllama(resolvedModel, fullPrompt)
}

type appLLMCaller struct {
	app *App
}

func (c *appLLMCaller) Chat(model, system, user string) (string, error) {
	fullPrompt := system + "\n\n" + user
	res := c.app.ChatWithModel(model, fullPrompt)
	return res, nil
}

func (a *App) StartUpgradePipeline(task string) {
	logging.Userf("StartUpgradePipeline task=%q", task)

	workspacePath := filepath.Join(os.TempDir(), "cria_workspace")

	if _, err := os.Stat(filepath.Join(workspacePath, ".git")); os.IsNotExist(err) {
		logging.Statef("workspace not found. extracting embedded source to %s", workspacePath)
		err := vcs.SetupWorkspaceFromZip(sourceZip, workspacePath)
		if err != nil {
			logging.Errorf("SetupWorkspaceFromZip: %v", err)
			return
		}
	}

	logging.Statef("workspace ready, launching orchestrator")

	models := a.GetOllamaModels()
	if len(models) == 0 {
		logging.Errorf("No Ollama models found. Cannot start pipeline.")
		runtime.EventsEmit(a.ctx, "pipeline-event", pipeline.PipelineEvent{
			Type:    "toast",
			Icon:    "❌",
			Content: "No Ollama models found. Please download a model in Settings first.",
		})
		return
	}
	defaultModel := models[0]
	logging.Statef("Using model: %s", defaultModel)

	llmCaller := &appLLMCaller{app: a}

	registry := pipeline.AgentRegistry{
		Architect:         agent.NewLLMArchitect(llmCaller, defaultModel),
		DesignCritic:      agent.NewMockDesignCritic(),
		UnitPlanner:       agent.NewMockGeneric("Unit Planner", "📅"),
		PlanCritic:        agent.NewMockGeneric("Plan Critic", "📋"),
		Developer:         agent.NewMockDeveloper(),
		CodeReviewer:      agent.NewMockCodeReviewer(),
		Tester:            agent.NewMockGeneric("Tester", "🧪"),
		TestVerifier:      agent.NewMockGeneric("Test Verifier", "✅"),
		Integrator:        agent.NewMockGeneric("Integrator", "🧩"),
		IntegrationCritic: agent.NewMockGeneric("Integration Critic", "⚖️"),
		FinalVerifier:     agent.NewMockFinalVerifier(),
		Watchdog:          agent.NewMockWatchdog(),
		Translator:        agent.NewMockTranslator(),
	}

	orc := pipeline.NewOrchestrator(a.ctx, workspacePath, registry)
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
		logging.Statef("ApplyUpgrade dev mode: simulating restart hash=%s version=%s", hash, version)
		CurrentCommit = hash
		CurrentVersion = version
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

	ldflags := fmt.Sprintf("-X main.CurrentCommit=%s -X main.CurrentVersion=%s", hash, version)
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

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
		logging.Statef("ollama runner killed for upgrade restart")
	}

	go func() {
		os.Exit(0)
	}()

	return nil
}
