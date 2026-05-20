package vcs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitManager struct {
	repoPath string
}

func NewGitManager(path string) *GitManager {
	return &GitManager{repoPath: path}
}

func (g *GitManager) execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, errBuf.String())
	}
	return strings.TrimSpace(outBuf.String()), nil
}

func (g *GitManager) InitIfNeeded() error {
	_, err := g.execGit("rev-parse", "--is-inside-work-tree")
	if err != nil {
		if _, err := g.execGit("init"); err != nil {
			return err
		}
		if _, err := g.execGit("add", "."); err != nil {
			return err
		}
		if _, err := g.execGit("commit", "-m", "Initial baseline"); err != nil {
			return err
		}
	}
	return nil
}

func (g *GitManager) StartUpgradeBranch(branchName string) error {
	g.execGit("checkout", "main")
	g.execGit("reset", "--hard")
	_, err := g.execGit("checkout", "-b", branchName)
	return err
}

func (g *GitManager) CommitAndMerge(branchName, commitMsg string) error {
	g.execGit("add", ".")
	g.execGit("commit", "-m", commitMsg)
	g.execGit("checkout", "main")
	_, err := g.execGit("merge", branchName)
	if err == nil {
		g.execGit("branch", "-D", branchName)
	}
	return err
}

func (g *GitManager) AbortBranch(branchName string) error {
	g.execGit("reset", "--hard")
	g.execGit("checkout", "main")
	_, err := g.execGit("branch", "-D", branchName)
	return err
}

func SetupShadowWorkspace(sourcePath, workspacePath string) error {
	_ = os.RemoveAll(workspacePath)
	cmd := exec.Command("git", "clone", sourcePath, workspacePath)
	return cmd.Run()
}
