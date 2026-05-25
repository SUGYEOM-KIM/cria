package pipeline

type Context struct {
	Task              string
	DesignSpecPath    string
	ErrorLogs         []string
	RetryCounts       map[string]int
	GlobalTransitions int
	CurrentStage      string
}

func NewContext(task string) *Context {
	return &Context{
		Task:         task,
		RetryCounts:  make(map[string]int),
		CurrentStage: "DESIGN",
	}
}
