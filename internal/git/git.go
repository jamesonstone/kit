// package git provides git integration for Kit.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsRepo checks if the given directory is inside a git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// CurrentBranch returns the name of the current git branch.
func CurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// BranchExists checks if a branch with the given name exists.
func BranchExists(dir string, branchName string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// CreateBranch creates a new branch from the base branch and checks it out.
func CreateBranch(dir string, branchName string, baseBranch string) error {
	// check if branch already exists
	if BranchExists(dir, branchName) {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}

	// create and checkout the new branch
	cmd := exec.Command("git", "checkout", "-b", branchName, baseBranch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch: %s", strings.TrimSpace(string(out)))
	}

	return nil
}

// CreateBranchFromCurrent creates a new branch from the current HEAD.
func CreateBranchFromCurrent(dir string, branchName string) error {
	// check if branch already exists
	if BranchExists(dir, branchName) {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}

	// create and checkout the new branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch: %s", strings.TrimSpace(string(out)))
	}

	return nil
}

// CheckoutBranch checks out an existing branch.
func CheckoutBranch(dir string, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// EnsureBranch creates a branch if it doesn't exist, or checks it out if it does.
func EnsureBranch(dir string, branchName string, baseBranch string) (created bool, err error) {
	if BranchExists(dir, branchName) {
		err = CheckoutBranch(dir, branchName)
		return false, err
	}

	if baseBranch != "" {
		err = CreateBranch(dir, branchName, baseBranch)
	} else {
		err = CreateBranchFromCurrent(dir, branchName)
	}
	return true, err
}

// HasUncommittedChanges checks if there are uncommitted changes.
func HasUncommittedChanges(dir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
