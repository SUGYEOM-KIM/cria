package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockArchitect struct{}

func NewMockArchitect() *MockArchitect { return &MockArchitect{} }
func (a *MockArchitect) Name() string  { return "Global Architect" }
func (a *MockArchitect) Icon() string  { return "📐" }

func (a *MockArchitect) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Drafting design_spec.md based on task..."})
	time.Sleep(1 * time.Second)

	return &pipeline.DesignResult{
		SpecPath:    "specs/design_spec.md",
		SpecContent: "# Architecture Spec\n...",
	}, nil
}
