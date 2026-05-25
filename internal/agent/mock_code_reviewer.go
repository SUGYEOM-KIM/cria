package agent

import (
	"context"
	"cria/internal/pipeline"
)

type MockCodeReviewer struct {
	rejectCount int
}

func NewMockCodeReviewer() *MockCodeReviewer { return &MockCodeReviewer{} }
func (a *MockCodeReviewer) Name() string     { return "Code Reviewer" }
func (a *MockCodeReviewer) Icon() string     { return "🔍" }

func (a *MockCodeReviewer) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	a.rejectCount++

	if a.rejectCount <= 3 {
		emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Syntax error detected. Rejecting code."})
		return false, nil
	}

	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Code review passed."})
	return true, nil
}
