package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockDesignCritic struct{}

func NewMockDesignCritic() *MockDesignCritic { return &MockDesignCritic{} }
func (a *MockDesignCritic) Name() string     { return "Design Critic" }
func (a *MockDesignCritic) Icon() string     { return "🧐" }

func (a *MockDesignCritic) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Verifying design constraints..."})
	time.Sleep(1 * time.Second)

	return true, nil
}
