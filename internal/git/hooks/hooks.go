package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/gitlab-org/gitaly/internal/config"
)

// HookDir is the last element for the hooks path
const HookDir = "hooks"

// SetGitLabHooks sets the required hooks so each Git operation performs
// authentication and authorization using the internal API of GitLab.
// This function is meant to be idempotent, which means that setting the hooks
// which are properly set already does not return an error
// Custom errors are returned to avoid leaking storage details
func SetGitLabHooks(repoPath string) error {
	fullPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("Unable to get full path for: %s", repoPath)
	}

	newDir := filepath.Join(fullPath, HookDir)

	// If the hooks dir, is actually a dir
	if fi, err := os.Stat(newDir); err == nil {
		if fi.IsDir() && rename(newDir) != nil {
			return fmt.Errorf("rename of existing hooks dir failed")
		}

		return link(newDir)
	}

	linkInfo, err := os.Readlink(newDir)
	if err != nil {
		if err := link(newDir); err != nil {
			return fmt.Errorf("unable to set non existing symlink for %s", repoPath)
		}
	}

	linkFullPath, err := filepath.Abs(linkInfo)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(linkFullPath, Path()) { // Old links are removed and relinked
		if err = rename(newDir); err != nil {
			return fmt.Errorf("unable to remove outdated hooks")
		}

		if err := link(newDir); err != nil {
			return fmt.Errorf("unable to set updated symlink for %s", repoPath)
		}
	}

	return nil
}

// Path is the full path to the location of the global hooks
func Path() string {
	fullPath, _ := filepath.Abs(filepath.Join(config.Config.GitlabShell.Dir, HookDir))
	return fullPath
}

func rename(oldPath string) error {
	renamePath := fmt.Sprintf("%s+%v", oldPath, time.Now().UnixNano())
	return os.Rename(oldPath, renamePath)
}

func link(newDir string) error {
	return os.Symlink(Path(), newDir)
}
