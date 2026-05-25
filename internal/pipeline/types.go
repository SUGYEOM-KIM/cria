package pipeline

type PipelineEvent struct {
	Type    string            `json:"type"`
	Icon    string            `json:"icon,omitempty"`
	Role    string            `json:"role,omitempty"`
	Action  string            `json:"action,omitempty"`
	Content string            `json:"content"`
	Data    map[string]string `json:"data,omitempty"`
}

type HITLResponse struct {
	Approved bool
	Feedback string
}

type DesignResult struct {
	SpecPath    string
	SpecContent string
	Summary     string
}

type ImplResult struct {
	FilesChanged []string
	Summary      string
}

type BumpType string

const (
	BumpMajor BumpType = "major"
	BumpMinor BumpType = "minor"
	BumpPatch BumpType = "patch"
)

type ReleaseResult struct {
	Bump       BumpType
	NewVersion string
	CommitMsg  string
}
