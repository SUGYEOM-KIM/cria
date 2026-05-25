package pipeline

import "context"

type LLMCaller interface {
	Chat(model, system, user string) (string, error)
}

type AgentInput struct {
	Task      string
	Workspace string
	Feedback  string
	Design    *DesignResult
	Impl      *ImplResult
	ErrorLogs []string
}

type Agent interface {
	Name() string
	Icon() string
	Run(ctx context.Context, in AgentInput, emit func(PipelineEvent)) (any, error)
}

type AgentRegistry struct {
	Architect         Agent
	DesignCritic      Agent
	UnitPlanner       Agent
	PlanCritic        Agent
	Developer         Agent
	CodeReviewer      Agent
	Tester            Agent
	TestVerifier      Agent
	Integrator        Agent
	IntegrationCritic Agent
	FinalVerifier     Agent
	Watchdog          Agent
	Translator        Agent
}
