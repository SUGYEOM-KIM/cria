package pipeline

import (
	"context"
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
	fmt.Printf("[PIPELINE] Starting mock pipeline for task: %s\n", task)
	branchName := fmt.Sprintf("upgrade-%d", time.Now().Unix())

	err := o.git.StartUpgradeBranch(branchName)
	if err != nil {
		o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Git setup failed"})
		return
	}

	o.Emit(PipelineEvent{Type: "toast", Icon: "🚀", Content: "System upgrade pipeline started!"})

	feedbackContext := ""

designLoop:
	for {
		o.Emit(PipelineEvent{Type: "status", Content: "Running DESIGN..."})
		time.Sleep(1 * time.Second)

		if feedbackContext != "" {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "Redesigning based on feedback: " + feedbackContext})
		} else {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "Analyzing task: " + task})
		}

		time.Sleep(2 * time.Second)
		o.Emit(PipelineEvent{Type: "system_msg", Icon: "🔧", Role: "Global Architect", Action: "WRITE", Content: "specs/design_spec.md design complete"})

		o.Emit(PipelineEvent{Type: "hitl", Content: "Design spec draft complete. Please review."})
		response := <-hitlChan

		if response.Approved {
			o.Emit(PipelineEvent{Type: "toast", Icon: "👍", Content: "Design approved. Moving to implementation."})
			break designLoop
		} else {
			o.Emit(PipelineEvent{Type: "toast", Icon: "🔁", Content: "Feedback received. Retrying design."})
			feedbackContext = response.Feedback
		}
	}

	o.Emit(PipelineEvent{Type: "status", Content: "Running IMPLEMENTATION..."})
	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "system_msg", Icon: "💻", Role: "Module Writer", Action: "EDIT_FILE", Content: "Writing implementation files..."})

	testFilePath := filepath.Join(o.workspacePath, "cria_test_log.txt")
	testContent := fmt.Sprintf("Task executed: %s\nTime: %v\n", task, time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(testFilePath, []byte(testContent), 0644)

	o.Emit(PipelineEvent{Type: "status", Content: "Running RELEASE AGENT..."})
	time.Sleep(1 * time.Second)

	currentVersion := o.git.GetLatestTag()
	if currentVersion == "" {
		currentVersion = "v0.0.0"
	}

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

	commitTitle := fmt.Sprintf("Auto-upgrade: %s: %s", prefix, task)
	commitBody := fmt.Sprintf("Implementation Details:\n- Processed objective: %s\n- Applied %s version bump\n\n[Auto-upgrade]", task, strings.ToUpper(bumpType))
	aiCommitMessage := fmt.Sprintf("%s\n\n%s", commitTitle, commitBody)

	o.Emit(PipelineEvent{Type: "system_msg", Icon: "🏷️", Role: "Release Manager", Content: fmt.Sprintf("Decision: %s bump. Tagging as %s", strings.ToUpper(bumpType), newVersion)})
	time.Sleep(1 * time.Second)

	o.Emit(PipelineEvent{Type: "system_msg", Icon: "📝", Role: "Release Manager", Content: "Generated human-readable commit message based on task analysis."})

	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "status", Content: "MERGING AND TAGGING..."})

	err = o.git.CommitAndMerge(branchName, aiCommitMessage)
	if err != nil {
		o.Emit(PipelineEvent{Type: "toast", Icon: "❌", Content: "Merge failed"})
		return
	}

	o.git.CreateTag(newVersion)

	o.Emit(PipelineEvent{Type: "toast", Icon: "✅", Content: fmt.Sprintf("Mission Complete! Version %s released.", newVersion)})
	o.Emit(PipelineEvent{Type: "complete", Content: ""})
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
