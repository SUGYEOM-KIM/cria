package pipeline

import (
	"context"
	"cria/internal/logging"
	"cria/internal/vcs"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Orchestrator struct {
	ctx           context.Context
	workspacePath string
	git           *vcs.GitManager
	agents        AgentRegistry
}

func NewOrchestrator(ctx context.Context, workspacePath string, registry AgentRegistry) *Orchestrator {
	return &Orchestrator{
		ctx:           ctx,
		workspacePath: workspacePath,
		git:           vcs.NewGitManager(workspacePath),
		agents:        registry,
	}
}

func (o *Orchestrator) Emit(event PipelineEvent) {
	runtime.EventsEmit(o.ctx, "pipeline-event", event)
}

func (o *Orchestrator) RunMock(task string, hitlChan chan HITLResponse) {
	if err := o.git.CheckoutUpgradeBranch(); err != nil {
		o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Git setup failed"})
		return
	}

	o.Emit(PipelineEvent{Type: "toast", Icon: "🚀", Content: "System upgrade pipeline started!"})

	pctx := NewContext(task)
	agentInput := AgentInput{
		Task:      task,
		Workspace: o.workspacePath,
	}

	for pctx.GlobalTransitions < 15 {
		pctx.GlobalTransitions++
		time.Sleep(500 * time.Millisecond)

		switch pctx.CurrentStage {
		case "DESIGN":
			o.Emit(PipelineEvent{Type: "status", Content: "RUNNING DESIGN LOOP..."})

			resArch, _ := o.agents.Architect.Run(o.ctx, agentInput, o.Emit)
			agentInput.Design = resArch.(*DesignResult)

			if err := o.writeDesignSpec(agentInput.Design); err != nil {
				logging.Errorf("orchestrator writeDesignSpec: %v", err)
				o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Failed to write design spec"})
				return
			}
			logging.Statef("orchestrator design spec written path=%s len=%d", agentInput.Design.SpecPath, len(agentInput.Design.SpecContent))

			resCritic, _ := o.agents.DesignCritic.Run(o.ctx, agentInput, o.Emit)
			approved := resCritic.(bool)

			if approved {
				o.Emit(PipelineEvent{
					Type:    "hitl",
					Content: "Design draft complete. Please review.",
					Data: map[string]string{
						"spec_path":    agentInput.Design.SpecPath,
						"spec_content": agentInput.Design.SpecContent,
					},
				})
				response := <-hitlChan

				if response.Approved {
					pctx.CurrentStage = "IMPLEMENTATION"
					agentInput.Feedback = ""
				} else {
					agentInput.Feedback = response.Feedback
					pctx.RetryCounts["DESIGN"]++
				}
			} else {
				pctx.RetryCounts["DESIGN"]++
				if pctx.RetryCounts["DESIGN"] >= 3 {
					pctx.CurrentStage = "WATCHDOG"
					agentInput.ErrorLogs = append(agentInput.ErrorLogs, "DESIGN stage recurring failures")
				}
			}

		case "IMPLEMENTATION":
			o.Emit(PipelineEvent{Type: "status", Content: "RUNNING IMPLEMENTATION LOOP..."})

			resDev, _ := o.agents.Developer.Run(o.ctx, agentInput, o.Emit)
			agentInput.Impl = resDev.(*ImplResult)

			resRev, _ := o.agents.CodeReviewer.Run(o.ctx, agentInput, o.Emit)
			approved := resRev.(bool)

			if approved {
				pctx.CurrentStage = "INTEGRATION"
			} else {
				pctx.RetryCounts["IMPLEMENTATION"]++
				agentInput.ErrorLogs = append(agentInput.ErrorLogs, "IMPLEMENTATION syntax error")

				if pctx.RetryCounts["IMPLEMENTATION"] >= 3 {
					pctx.CurrentStage = "WATCHDOG"
				}
			}

		case "WATCHDOG":
			o.Emit(PipelineEvent{Type: "status", Content: "WATCHDOG INTERVENTION"})

			resWatch, _ := o.agents.Watchdog.Run(o.ctx, agentInput, o.Emit)
			targetStage := resWatch.(string)

			resTrans, _ := o.agents.Translator.Run(o.ctx, AgentInput{Feedback: targetStage}, o.Emit)
			o.Emit(PipelineEvent{Type: "toast", Icon: "🚨", Content: resTrans.(string)})

			pctx.RetryCounts["DESIGN"] = 0
			pctx.RetryCounts["IMPLEMENTATION"] = 0
			agentInput.ErrorLogs = nil
			pctx.CurrentStage = targetStage

		case "INTEGRATION":
			o.Emit(PipelineEvent{Type: "status", Content: "RUNNING INTEGRATION LOOP..."})

			_, _ = o.agents.FinalVerifier.Run(o.ctx, agentInput, o.Emit)

			o.Emit(PipelineEvent{Type: "status", Content: "COMMITTING TO cria-update..."})
			_ = o.git.CommitOnBranch("Auto-upgrade complete")

			currentVersion := o.git.GetLatestTag()
			if currentVersion == "" {
				currentVersion = "v0.0.0"
			}
			newVersion := bumpVersion(currentVersion, "minor")
			_ = o.git.CreateTag(newVersion)

			o.Emit(PipelineEvent{Type: "toast", Icon: "✅", Content: fmt.Sprintf("Mission Complete! %s released.", newVersion)})
			o.Emit(PipelineEvent{Type: "complete", Content: ""})
			return
		}
	}

	o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Global Timeout Reached. System Abort."})
}

func (o *Orchestrator) writeDesignSpec(d *DesignResult) error {
	if d == nil || d.SpecPath == "" {
		return fmt.Errorf("design result missing path")
	}
	abs := filepath.Join(o.workspacePath, d.SpecPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(d.SpecContent), 0644); err != nil {
		return fmt.Errorf("write %s: %w", abs, err)
	}
	return nil
}

func bumpVersion(current string, bumpType string) string {
	current = strings.TrimPrefix(current, "v")
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		parts = []string{"0", "0", "0"}
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	default:
		patch++
	}
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
