package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockFinalVerifier struct{}

func NewMockFinalVerifier() *MockFinalVerifier { return &MockFinalVerifier{} }
func (a *MockFinalVerifier) Name() string      { return "Final Verifier" }
func (a *MockFinalVerifier) Icon() string      { return "🏆" }

func (a *MockFinalVerifier) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Running end-to-end integration tests..."})
	time.Sleep(1 * time.Second)

	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "All integration tests passed. System is stable."})
	return true, nil
}
