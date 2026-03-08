package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitStatus represents the git status of a file.
type GitStatus int

const (
	GitNone       GitStatus = iota
	GitModified             // Modified in worktree (not staged)
	GitStaged               // Staged for commit
	GitUntracked            // Not tracked by git
	GitConflict             // Merge conflict
	GitRenamed              // Renamed
	GitDeleted              // Deleted
	GitIgnored              // Ignored by .gitignore
	GitAdded                // Newly added and staged
)

// GitInfo holds git repository information for the current directory.
type GitInfo struct {
	IsRepo     bool
	Branch     string
	FileStatus map[string]GitStatus // relative path -> status
	HasChanges bool                 // any uncommitted changes
}

// gitBranchIcon is the Nerd Font icon for git branch.
const gitBranchIcon = "\U000f062c" // nf-md-source_branch

// gitStatusIcons maps git status to display characters.
var gitStatusIcons = map[GitStatus]string{
	GitModified:  "M",
	GitStaged:    "S",
	GitUntracked: "?",
	GitConflict:  "!",
	GitRenamed:   "R",
	GitDeleted:   "D",
	GitAdded:     "A",
}

// detectGitInfo gathers git branch and file status information for the
// given directory. Returns a zero-value GitInfo if not in a git repo.
func detectGitInfo(dir string) GitInfo {
	info := GitInfo{
		FileStatus: make(map[string]GitStatus),
	}

	// Check if we're in a git repo and get branch name
	branch, err := gitBranch(dir)
	if err != nil {
		return info
	}
	info.IsRepo = true
	info.Branch = branch

	// Get file statuses
	info.FileStatus, info.HasChanges = gitFileStatuses(dir)

	return info
}

// gitBranch returns the current branch name or HEAD ref.
func gitBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		// Detached HEAD — get short SHA
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd.Dir = dir
		out, err = cmd.Output()
		if err != nil {
			return "HEAD", nil
		}
		return strings.TrimSpace(string(out)), nil
	}
	return branch, nil
}

// gitFileStatuses returns a map of file paths (relative to dir) to their
// git status, and whether there are any changes at all.
func gitFileStatuses(dir string) (map[string]GitStatus, bool) {
	statuses := make(map[string]GitStatus)

	cmd := exec.Command("git", "status", "--porcelain=v1", "-uall")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return statuses, false
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return statuses, false
	}

	// Get the repo root so we can compute paths relative to dir
	repoRoot := gitRepoRoot(dir)

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		x := line[0] // index (staged) status
		y := line[1] // worktree status
		path := line[3:]

		// Handle renames: "R  old -> new"
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[idx+4:]
		}

		status := parseGitXY(x, y)
		if status == GitNone {
			continue
		}

		// Convert path to be relative to the current directory
		absPath := filepath.Join(repoRoot, path)
		relPath, err := filepath.Rel(dir, absPath)
		if err != nil {
			continue
		}

		// Store status for the file itself
		statuses[relPath] = status

		// Also store status for parent directories so directory entries
		// show indicators when they contain modified files.
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 1 {
			// The first component is the immediate child dir
			statuses[parts[0]] = mergeGitStatus(statuses[parts[0]], status)
		}
	}

	return statuses, len(statuses) > 0
}

// gitRepoRoot returns the root of the git repository.
func gitRepoRoot(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	return strings.TrimSpace(string(out))
}

// parseGitXY converts the two-character status from git status --porcelain
// into a GitStatus. The X column is the index (staged) status, Y is the
// worktree status.
func parseGitXY(x, y byte) GitStatus {
	// Unmerged (conflict) takes priority
	if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
		return GitConflict
	}

	// Untracked
	if x == '?' && y == '?' {
		return GitUntracked
	}

	// Ignored
	if x == '!' && y == '!' {
		return GitIgnored
	}

	// Worktree modifications take display priority over staged
	switch y {
	case 'M':
		return GitModified
	case 'D':
		return GitDeleted
	}

	// Staged changes
	switch x {
	case 'M':
		return GitStaged
	case 'A':
		return GitAdded
	case 'D':
		return GitDeleted
	case 'R':
		return GitRenamed
	}

	return GitNone
}

// mergeGitStatus returns the "more important" status for directory
// aggregation. Modified > Untracked > Staged > others.
func mergeGitStatus(existing, incoming GitStatus) GitStatus {
	if existing == GitNone {
		return incoming
	}
	// Conflict is highest priority
	if existing == GitConflict || incoming == GitConflict {
		return GitConflict
	}
	// Modified next
	if existing == GitModified || incoming == GitModified {
		return GitModified
	}
	// Untracked
	if existing == GitUntracked || incoming == GitUntracked {
		return GitUntracked
	}
	return existing
}

// lookupGitStatus returns the git status for a given entry name.
func lookupGitStatus(info *GitInfo, name string) GitStatus {
	if info == nil || !info.IsRepo {
		return GitNone
	}
	if s, ok := info.FileStatus[name]; ok {
		return s
	}
	return GitNone
}
