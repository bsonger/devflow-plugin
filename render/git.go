package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitCloneOrPull clones a repo if baseDir does not exist, otherwise pulls the latest changes.
func GitCloneOrPull(repoURL, branch, baseDir string) error {
	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(baseDir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("Cloning %s (branch: %s) into %s...\n", repoURL, branch, baseDir)
		cmd := exec.Command("git", "-c", "http.version=HTTP/1.1", "clone", "-b", branch, repoURL, baseDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to git clone: %w", err)
		}
	} else {
		fmt.Printf("Pulling latest changes for %s (branch: %s)...\n", baseDir, branch)
		cmd := exec.Command("git", "-C", baseDir, "-c", "http.version=HTTP/1.1", "pull", "origin", branch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to git pull: %w", err)
		}
	}

	return nil
}

func RepoDirName(repoURL string) string {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "-" + parts[len(parts)-1]
	}
	return strings.ReplaceAll(repoURL, "/", "_")
}
