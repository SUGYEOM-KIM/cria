package pipeline

import (
	"context"
	"cria/internal/vcs"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type PipelineEvent struct {
	Type    string      `json:"type"`
	Icon    string      `json:"icon,omitempty"`
	Role    string      `json:"role,omitempty"`
	Content string      `json:"content,omitempty"`
	Action  string      `json:"action,omitempty"`
	Params  interface{} `json:"params,omitempty"`
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
	gitMgr := vcs.NewGitManager(workspacePath)
	gitMgr.InitIfNeeded()

	return &Orchestrator{
		ctx:           ctx,
		workspacePath: workspacePath,
		git:           gitMgr,
	}
}

func (o *Orchestrator) Emit(event PipelineEvent) {
	runtime.EventsEmit(o.ctx, "pipeline-event", event)
}

func (o *Orchestrator) RunMock(task string, hitlChan chan HITLResponse) {
	branchName := fmt.Sprintf("upgrade-%d", time.Now().Unix())
	o.git.StartUpgradeBranch(branchName)

	o.Emit(PipelineEvent{Type: "toast", Icon: "🚀", Content: "System upgrade pipeline 시작!"})

	feedbackContext := ""

designLoop:
	for {
		o.Emit(PipelineEvent{Type: "status", Content: "Running DESIGN..."})
		time.Sleep(1 * time.Second)

		if feedbackContext != "" {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "피드백을 반영하여 재설계 중: " + feedbackContext})
		} else {
			o.Emit(PipelineEvent{Type: "system_msg", Icon: "📐", Role: "Global Architect", Content: "요청 분석 중: " + task})
		}

		time.Sleep(2 * time.Second)
		o.Emit(PipelineEvent{Type: "system_msg", Icon: "🔧", Role: "Global Architect", Action: "WRITE", Content: "specs/design_spec.md 설계 완료"})

		time.Sleep(1 * time.Second)
		o.Emit(PipelineEvent{Type: "hitl", Content: "Architectural design spec draft is complete. Please review the spec and click Approve, or provide feedback to retry."})

		response := <-hitlChan

		if response.Approved {
			o.Emit(PipelineEvent{Type: "toast", Icon: "👍", Content: "사용자가 설계를 승인했습니다. 구현 단계로 넘어갑니다."})
			break designLoop
		} else {
			o.Emit(PipelineEvent{Type: "toast", Icon: "🔁", Content: "피드백이 전달되었습니다. 설계를 다시 진행합니다."})
			feedbackContext = response.Feedback
		}
	}

	o.Emit(PipelineEvent{Type: "status", Content: "Running IMPLEMENTATION..."})

	time.Sleep(2 * time.Second)
	o.Emit(PipelineEvent{Type: "system_msg", Icon: "💻", Role: "Module Writer", Action: "EDIT_FILE", Content: "core/llm_provider.go 기능 패치 적용"})

	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "system_msg", Icon: "🧪", Role: "Tester", Action: "WRITE", Content: "tests/test_llm_provider.go 단위 테스트 작성 완료"})

	time.Sleep(2 * time.Second)
	o.Emit(PipelineEvent{Type: "status", Content: "Running INTEGRATION..."})

	time.Sleep(2 * time.Second)
	o.Emit(PipelineEvent{Type: "system_msg", Icon: "✅", Role: "Test Verifier", Action: "RUN_PYTEST", Content: "모든 테스트 통과 (12 passed)"})

	o.git.CommitAndMerge(branchName, fmt.Sprintf("Auto-upgrade: %s", task))

	time.Sleep(1 * time.Second)
	o.Emit(PipelineEvent{Type: "toast", Icon: "🎊", Content: "Mission Complete! 안전하게 메인 코드에 병합되었습니다."})
}
