package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockGeneric struct {
	name string
	icon string
}

func NewMockGeneric(name, icon string) *MockGeneric {
	return &MockGeneric{name: name, icon: icon}
}

func (a *MockGeneric) Name() string { return a.name }
func (a *MockGeneric) Icon() string { return a.icon }

func (a *MockGeneric) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: "Running " + a.name + " tasks..."})
	time.Sleep(500 * time.Millisecond)

	return true, nil
}
