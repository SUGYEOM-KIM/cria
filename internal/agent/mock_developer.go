package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockDeveloper struct {
	failCount int
}

func NewMockDeveloper() *MockDeveloper { return &MockDeveloper{} }
func (a *MockDeveloper) Name() string  { return "Developer" }
func (a *MockDeveloper) Icon() string  { return "💻" }

func (a *MockDeveloper) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	a.failCount++
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Action: "WRITE", Content: "Implementing code modules..."})
	time.Sleep(1 * time.Second)

	return &pipeline.ImplResult{
		FilesChanged: []string{"main.go"},
	}, nil
}
