package agent

import (
	"context"
	"cria/internal/pipeline"
)

type MockTranslator struct{}

func NewMockTranslator() *MockTranslator { return &MockTranslator{} }
func (a *MockTranslator) Name() string   { return "Translator" }
func (a *MockTranslator) Icon() string   { return "🌐" }

func (a *MockTranslator) Run(ctx context.Context, in pipeline.AgentInput, emit func(pipeline.PipelineEvent)) (any, error) {
	translatedMsg := "설계 규격에서 치명적인 모순이 발견되어 " + in.Feedback + " 단계로 롤백합니다."
	emit(pipeline.PipelineEvent{Type: "system_msg", Icon: a.Icon(), Role: a.Name(), Content: translatedMsg})

	return translatedMsg, nil
}
