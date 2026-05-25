package tool

import "context"

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args string) (string, error)
}

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string        { return "READ_FILE" }
func (t *ReadFileTool) Description() string { return "Reads a file content" }
func (t *ReadFileTool) Execute(ctx context.Context, args string) (string, error) {
	return "mock content", nil
}

type SyncPackageTool struct{}

func (t *SyncPackageTool) Name() string        { return "RUN_PACKAGE_MANAGER" }
func (t *SyncPackageTool) Description() string { return "Syncs project dependencies" }
func (t *SyncPackageTool) Execute(ctx context.Context, args string) (string, error) {
	return "dependencies synced", nil
}
