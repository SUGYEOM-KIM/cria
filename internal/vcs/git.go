package vcs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

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
	g.execGit("config", "user.name", "Cria Agent")
	g.execGit("config", "user.email", "cria@agent.local")

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
	rawBranch, err := g.execGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	currentBranch := strings.TrimSpace(rawBranch)

	if _, err := g.execGit("checkout", currentBranch); err != nil {
		return err
	}

	if _, err := g.execGit("reset", "--hard"); err != nil {
		return err
	}

	if _, err := g.execGit("checkout", "-b", branchName); err != nil {
		return err
	}
	return nil
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

type UpgradeHistory struct {
	Version       string `json:"version"`
	Hash          string `json:"hash"`
	Message       string `json:"message"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	IsAutoUpgrade bool   `json:"isAutoUpgrade"`
}

func (g *GitManager) RollbackToHash(hash string) error {
	targetHash := hash + "^"

	_, err := g.execGit("rev-parse", "--verify", targetHash)
	if err != nil {
		_, err = g.execGit("update-ref", "-d", "HEAD")
		if err != nil {
			return fmt.Errorf("failed to delete HEAD ref: %v", err)
		}
		g.execGit("read-tree", "--empty")
		g.execGit("clean", "-fdx")
		return nil
	}

	if _, err := g.execGit("reset", "--hard", targetHash); err != nil {
		return fmt.Errorf("git reset failed: %v", err)
	}

	g.execGit("clean", "-fd")
	return nil
}

func (g *GitManager) GetUpgradeHistory() ([]UpgradeHistory, error) {
	out, err := g.execGit("log", "--pretty=format:%H|%s|%cd|%D", "--date=format:%Y-%m-%d|%H:%M:%S", "-n", "30")
	if err != nil {
		return nil, err
	}

	var history []UpgradeHistory
	if out == "" {
		return history, nil
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			version := ""
			if len(parts) >= 5 {
				decorations := parts[4]
				if strings.Contains(decorations, "tag: ") {
					tagPart := strings.Split(decorations, "tag: ")[1]
					version = strings.Split(tagPart, ",")[0]
					version = strings.TrimSpace(version)
				}
			}

			rawMsg := parts[1]
			cleanMsg := rawMsg
			isAuto := false

			if strings.HasPrefix(rawMsg, "Auto-upgrade:") {
				isAuto = true
				cleanMsg = strings.TrimSpace(strings.TrimPrefix(rawMsg, "Auto-upgrade:"))
			} else if strings.Contains(rawMsg, "Initial baseline") {
				isAuto = true
				if version == "" {
					version = "v0.0.0"
				}
			}

			history = append(history, UpgradeHistory{
				Hash:          parts[0],
				Message:       cleanMsg,
				Date:          parts[2],
				Time:          parts[3],
				Version:       version,
				IsAutoUpgrade: isAuto,
			})
		}
	}
	return history, nil
}

func (g *GitManager) GetLatestTag() string {
	out, err := g.execGit("describe", "--tags", "--abbrev=0")
	if err != nil || out == "" {
		return "v0.0.0"
	}
	return strings.TrimSpace(out)
}

func (g *GitManager) CreateTag(version string) error {
	_, err := g.execGit("tag", "-f", version)
	return err
}

func (g *GitManager) CommitAndMerge(branchName, commitMsg string) error {
	if _, err := g.execGit("add", "."); err != nil {
		return fmt.Errorf("git add failed: %v", err)
	}

	status, _ := g.execGit("status", "--porcelain")
	if status != "" {
		if _, err := g.execGit("commit", "-m", commitMsg); err != nil {
			return fmt.Errorf("git commit failed: %v", err)
		}
	}

	targetBranch := "main"
	if _, err := g.execGit("show-ref", "--verify", "refs/heads/main"); err != nil {
		targetBranch = "master"
	}

	if _, err := g.execGit("checkout", targetBranch); err != nil {
		return fmt.Errorf("checkout %s failed: %v", targetBranch, err)
	}

	if _, err := g.execGit("merge", branchName); err != nil {
		return fmt.Errorf("merge failed: %v", err)
	}

	g.execGit("branch", "-D", branchName)
	return nil
}

func (g *GitManager) GetRootCommitHash() (string, error) {
	return g.execGit("rev-list", "--max-parents=0", "HEAD")
}
