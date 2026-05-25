package agent

import (
	"context"
	"cria/internal/pipeline"
	"fmt"
	"time"
)

type LLMArchitect struct {
	llm   pipeline.LLMCaller
	model string
}

func NewLLMArchitect(llm pipeline.LLMCaller, model string) *LLMArchitect {
	return &LLMArchitect{
		llm:   llm,
		model: model,
	}
}

func (a *LLMArchitect) Name() string { return "Global Architect" }
func (a *LLMArchitect) Icon() string { return "📐" }

func (a *LLMArchitect) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "status", Content: "ARCHITECT IS THINKING..."})

	systemPrompt := `You are the Global Architect of an AI development team. 
Your job is to create a comprehensive software architecture design based on the user's task.
Respond ONLY with a valid Markdown document containing the design specs.`

	userPrompt := fmt.Sprintf("Task: %s\nWorkspace: %s\n\nPlease provide the design_spec.md content.", in.Task, in.Workspace)

	if in.Feedback != "" {
		emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "피드백을 반영하여 재설계합니다: " + in.Feedback})
		userPrompt += fmt.Sprintf("\n\n[REVISION REQUIRED]: %s", in.Feedback)
	} else {
		emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "요구사항 분석 및 아키텍처 설계를 시작합니다..."})
	}

	start := time.Now()
	response, err := a.llm.Chat(a.model, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("llm chat failed: %w", err)
	}
	elapsed := time.Since(start)

	emit(pipeline.PipelineEvent{
		Type:    "system_msg",
		Icon:    a.Icon(),
		Role:    a.Name(),
		Action:  "WRITE",
		Content: fmt.Sprintf("설계 완료 (소요시간: %v). specs/design_spec.md 문서를 생성했습니다.", elapsed.Round(time.Second)),
	})

	return &pipeline.DesignResult{
		SpecPath:    "specs/design_spec.md",
		SpecContent: response,
		Summary:     "Architecture design completed.",
	}, nil
}
