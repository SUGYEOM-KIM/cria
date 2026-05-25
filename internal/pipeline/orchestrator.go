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

type PipelineEvent struct {
	Type    string `json:"type"`
	Icon    string `json:"icon,omitempty"`
	Role    string `json:"role,omitempty"`
	Action  string `json:"action,omitempty"`
	Content string `json:"content"`
}

type HITLResponse struct {
	Approved bool
	Feedback string
}

type Orchestrator struct {
	ctx           context.Context
	workspacePath string
	git           *vcs.GitManager
}

func NewOrchestrator(ctx context.Context, workspacePath string) *Orchestrator {
	return &Orchestrator{
		ctx:           ctx,
		workspacePath: workspacePath,
		git:           vcs.NewGitManager(workspacePath),
	}
}

func (o *Orchestrator) Emit(event PipelineEvent) {
	runtime.EventsEmit(o.ctx, "pipeline-event", event)
}

func (o *Orchestrator) RunMock(task string, hitlChan chan HITLResponse) {
	logging.Statef("pipeline.RunMock start task=%q workspace=%s", task, o.workspacePath)

	err := o.git.CheckoutUpgradeBranch()
	if err != nil {
		logging.Errorf("pipeline CheckoutUpgradeBranch: %v", err)
		o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Git setup failed"})
		return
	}
	logging.Statef("pipeline on cria-update branch")

	o.Emit(PipelineEvent{Type: "toast", Icon: "🚀", Content: "System upgrade pipeline started!"})

	feedbackContext := ""
	designIteration := 0

designLoop:
	for {
		designIteration++
		logging.Statef("pipeline DESIGN iteration=%d feedback=%q", designIteration, feedbackContext)
		o.Emit(PipelineEvent{Type: "status", Content: "Running DESIGN..."})
		time.Sleep(1 * time.Second)

		if feedbackContext != "" {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "Redesigning based on feedback: " + feedbackContext})
		} else {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "Analyzing task: " + task})
		}

		time.Sleep(2 * time.Second)
		o.Emit(PipelineEvent{Type: "system_msg", Icon: "🔧", Role: "Global Architect", Action: "WRITE", Content: "specs/design_spec.md design complete"})

		logging.Statef("pipeline waiting for HITL approval (iteration=%d)", designIteration)
		o.Emit(PipelineEvent{Type: "hitl", Content: "Design spec draft complete. Please review."})
		response := <-hitlChan
		logging.Statef("pipeline HITL response approved=%v feedback=%q", response.Approved, response.Feedback)

		if response.Approved {
			o.Emit(PipelineEvent{Type: "toast", Icon: "👍", Content: "Design approved. Moving to implementation."})
			break designLoop
		} else {
			o.Emit(PipelineEvent{Type: "toast", Icon: "🔁", Content: "Feedback received. Retrying design."})
			feedbackContext = response.Feedback
		}
	}

	logging.Statef("pipeline IMPLEMENTATION")
	o.Emit(PipelineEvent{Type: "status", Content: "Running IMPLEMENTATION..."})
	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "system_msg", Icon: "💻", Role: "Module Writer", Action: "EDIT_FILE", Content: "Writing implementation files..."})

	testFilePath := filepath.Join(o.workspacePath, "cria_test_log.txt")
	testContent := fmt.Sprintf("Task executed: %s\nTime: %v\n", task, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		logging.Errorf("pipeline write %s: %v", testFilePath, err)
	} else {
		logging.Statef("pipeline wrote %s", testFilePath)
	}

	logging.Statef("pipeline RELEASE")
	o.Emit(PipelineEvent{Type: "status", Content: "Running RELEASE AGENT..."})
	time.Sleep(1 * time.Second)

	currentVersion := o.git.GetLatestTag()
	logging.Statef("pipeline current version detected: %s", currentVersion)

	o.Emit(PipelineEvent{Type: "system_msg", Icon: "🤖", Role: "Release Manager", Content: fmt.Sprintf("Current version detected: %s", currentVersion)})
	time.Sleep(1 * time.Second)

	bumpType := "patch"
	prefix := "fix"
	taskLower := strings.ToLower(task)

	if strings.Contains(taskLower, "major") || strings.Contains(taskLower, "breaking") {
		bumpType = "major"
		prefix = "feat!"
	} else if strings.Contains(taskLower, "feature") || strings.Contains(taskLower, "minor") || strings.Contains(taskLower, "add") {
		bumpType = "minor"
		prefix = "feat"
	}

	newVersion := bumpVersion(currentVersion, bumpType)
	logging.Statef("pipeline bump=%s newVersion=%s prefix=%s", bumpType, newVersion, prefix)

	commitTitle := fmt.Sprintf("Auto-upgrade: %s: %s", prefix, task)
	commitBody := fmt.Sprintf("Implementation Details:\n- Processed objective: %s\n- Applied %s version bump\n\n[Auto-upgrade]", task, strings.ToUpper(bumpType))
	aiCommitMessage := fmt.Sprintf("%s\n\n%s", commitTitle, commitBody)

	o.Emit(PipelineEvent{Type: "system_msg", Icon: "🏷️", Role: "Release Manager", Content: fmt.Sprintf("Decision: %s bump. Tagging as %s", strings.ToUpper(bumpType), newVersion)})
	time.Sleep(1 * time.Second)

	o.Emit(PipelineEvent{Type: "system_msg", Icon: "📝", Role: "Release Manager", Content: "Generated human-readable commit message based on task analysis."})

	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "status", Content: "COMMITTING TO cria-update..."})

	err = o.git.CommitOnBranch(aiCommitMessage)
	if err != nil {
		logging.Errorf("pipeline CommitOnBranch: %v", err)
		o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Commit failed"})
		return
	}
	logging.Statef("pipeline commit done")

	if err := o.git.CreateTag(newVersion); err != nil {
		logging.Errorf("pipeline CreateTag %s: %v", newVersion, err)
	} else {
		logging.Statef("pipeline tag created: %s", newVersion)
	}

	o.Emit(PipelineEvent{Type: "toast", Icon: "✅", Content: fmt.Sprintf("Mission Complete! Version %s released.", newVersion)})
	o.Emit(PipelineEvent{Type: "complete", Content: ""})
	logging.Statef("pipeline.RunMock end version=%s", newVersion)
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
