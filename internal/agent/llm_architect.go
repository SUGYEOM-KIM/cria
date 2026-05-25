package agent

import (
	"context"
	"cria/internal/logging"
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
	logging.Statef("LLMArchitect.Run start model=%s task=%q feedback=%q", a.model, in.Task, in.Feedback)
	emit(pipeline.PipelineEvent{Type: "status", Content: "ARCHITECT IS THINKING..."})

	systemPrompt := `You are the Global Architect of an AI development team. 
Your job is to create a comprehensive software architecture design based on the user's task.
Respond ONLY with a valid Markdown document containing the design specs.`

	userPrompt := fmt.Sprintf("Task: %s\nWorkspace: %s\n\nPlease provide the design_spec.md content.", in.Task, in.Workspace)

	if in.Feedback != "" {
		logging.Statef("LLMArchitect revising based on feedback (len=%d)", len(in.Feedback))
		emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Revising the design based on your feedback: " + in.Feedback})
		userPrompt += fmt.Sprintf("\n\n[REVISION REQUIRED]: %s", in.Feedback)
	} else {
		logging.Statef("LLMArchitect drafting initial design")
		emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Analyzing requirements and drafting the architecture..."})
	}

	logging.Debugf("LLMArchitect prompt sizes: system=%d user=%d", len(systemPrompt), len(userPrompt))

	start := time.Now()
	response, err := a.llm.Chat(a.model, systemPrompt, userPrompt)
	elapsed := time.Since(start)
	if err != nil {
		logging.Errorf("LLMArchitect chat failed after %v: %v", elapsed.Round(time.Millisecond), err)
		return nil, fmt.Errorf("llm chat failed: %w", err)
	}
	logging.Statef("LLMArchitect chat completed in %v responseLen=%d", elapsed.Round(time.Millisecond), len(response))

	emit(pipeline.PipelineEvent{
		Type:    "system_msg",
		Icon:    a.Icon(),
		Role:    a.Name(),
		Action:  "WRITE",
		Content: fmt.Sprintf("Design complete (took %v). Generated specs/design_spec.md.", elapsed.Round(time.Second)),
	})

	logging.Statef("LLMArchitect.Run end")
	return &pipeline.DesignResult{
		SpecPath:    "specs/design_spec.md",
		SpecContent: response,
		Summary:     "Architecture design completed.",
	}, nil
}
