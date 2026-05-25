package agent

import (
	"context"
	"cria/internal/pipeline"
	"time"
)

type MockWatchdog struct{}

func NewMockWatchdog() *MockWatchdog { return &MockWatchdog{} }
func (a *MockWatchdog) Name() string { return "Watchdog" }
func (a *MockWatchdog) Icon() string { return "🚨" }

func (a *MockWatchdog) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Action: "DIAGNOSE", Content: "Analyzing multiple recurring failures..."})
	time.Sleep(2 * time.Second)

	return "DESIGN", nil
}
