package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Sandbox provides an isolated git worktree environment for task execution
type Sandbox struct {
	RepoPath     string // Original repository path
	WorktreePath string // Isolated worktree path
	Branch       string // Branch or commit to checkout
}

// NewSandbox creates a new git worktree sandbox
func NewSandbox(repoPath, initialState string) (*Sandbox, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repo path: %w", err)
	}

	// Verify it's a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = absPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", absPath)
	}

	// Create unique worktree path in $HOME/.cmb/cmb-worktrees/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Generate unique repo identifier (use basename of repo path)
	repoName := filepath.Base(absPath)

	// Create worktree path: $HOME/.cmb/cmb-worktrees/<repo-name>/run-<pid>
	worktreeBase := filepath.Join(homeDir, ".cmb", "cmb-worktrees", repoName)
	worktreePath := filepath.Join(worktreeBase, fmt.Sprintf("run-%d", os.Getpid()))

	// Ensure the base directory exists
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Determine what to checkout
	branch := initialState
	if branch == "" {
		// Get current branch
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = absPath
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		branch = strings.TrimSpace(string(output))
	}

	// Create worktree
	cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
	cmd.Dir = absPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w\n%s", err, output)
	}

	return &Sandbox{
		RepoPath:     absPath,
		WorktreePath: worktreePath,
		Branch:       branch,
	}, nil
}

// Cleanup removes the sandbox worktree
func (s *Sandbox) Cleanup() error {
	if s.WorktreePath == "" {
		return nil
	}

	// Remove worktree
	cmd := exec.Command("git", "worktree", "remove", s.WorktreePath, "--force")
	cmd.Dir = s.RepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try manual removal as fallback
		if removeErr := os.RemoveAll(s.WorktreePath); removeErr != nil {
			return fmt.Errorf("failed to remove worktree: %w\n%s", err, output)
		}
	}

	// Prune worktree metadata
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = s.RepoPath
	_ = cmd.Run() // Ignore errors for cleanup

	return nil
}

// GetDiff returns the git diff in the sandbox, including untracked files
func (s *Sandbox) GetDiff() (string, error) {
	var fullDiff strings.Builder

	// Get diff for tracked files that were modified
	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = s.WorktreePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	fullDiff.WriteString(string(output))

	// Get list of untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = s.WorktreePath
	untrackedOutput, err := cmd.Output()
	if err != nil {
		return fullDiff.String(), fmt.Errorf("failed to list untracked files: %w", err)
	}

	// For each untracked file, show it as a new file diff
	untrackedFiles := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
	for _, file := range untrackedFiles {
		if file == "" {
			continue
		}

		// Read the file content
		filePath := filepath.Join(s.WorktreePath, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		// Format as a diff for a new file
		fullDiff.WriteString(fmt.Sprintf("\ndiff --git a/%s b/%s\n", file, file))
		fullDiff.WriteString("new file mode 100644\n")
		fullDiff.WriteString("--- /dev/null\n")
		fullDiff.WriteString(fmt.Sprintf("+++ b/%s\n", file))
		fullDiff.WriteString("@@ -0,0 +1,")
		lines := strings.Split(string(content), "\n")
		fullDiff.WriteString(fmt.Sprintf("%d @@\n", len(lines)))
		for _, line := range lines {
			if line != "" || len(lines) > 1 {
				fullDiff.WriteString("+" + line + "\n")
			}
		}
	}

	return fullDiff.String(), nil
}

// GetStats returns statistics about changes in the sandbox
func (s *Sandbox) GetStats() (filesModified, linesAdded, linesDeleted int, err error) {
	// First, get stats for modified tracked files
	cmd := exec.Command("git", "diff", "--numstat", "HEAD")
	cmd.Dir = s.WorktreePath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get diff stats: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var added, deleted int
		var filename string
		if _, err := fmt.Sscanf(line, "%d\t%d\t%s", &added, &deleted, &filename); err == nil {
			filesModified++
			linesAdded += added
			linesDeleted += deleted
		}
	}

	// Also check for untracked files (new files created by the agent)
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = s.WorktreePath
	untrackedOutput, err := cmd.Output()
	if err != nil {
		return filesModified, linesAdded, linesDeleted, fmt.Errorf("failed to list untracked files: %w", err)
	}

	untrackedFiles := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
	for _, file := range untrackedFiles {
		if file == "" {
			continue
		}
		filesModified++

		// Count lines in the new file
		filePath := filepath.Join(s.WorktreePath, file)
		content, err := os.ReadFile(filePath)
		if err == nil {
			// Count non-empty lines
			lineCount := len(strings.Split(string(content), "\n"))
			if len(content) > 0 && content[len(content)-1] != '\n' {
				// File doesn't end with newline, but still counts as lines
				linesAdded += lineCount
			} else {
				linesAdded += lineCount - 1 // Don't count the final empty line
			}
		}
	}

	return filesModified, linesAdded, linesDeleted, nil
}
