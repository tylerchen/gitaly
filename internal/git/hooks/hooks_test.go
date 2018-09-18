package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/config"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestSetGitLabHooks(t *testing.T) {
	_, repoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	defer func(path string) { config.Config.GitlabShell.Dir = path }(config.Config.GitlabShell.Dir)
	tempHookDir, err := os.Getwd()
	require.NoError(t, err)

	hooksPath := filepath.Join(repoPath, HookDir)
	require.NoError(t, os.RemoveAll(hooksPath))

	// Tests initial setting, idempotent property, and updating to new location
	for _, newDir := range []string{tempHookDir, tempHookDir, "/tmp"} {
		config.Config.GitlabShell.Dir = newDir

		require.NoError(t, SetGitLabHooks(repoPath))

		_, err := os.Lstat(hooksPath)
		require.NoError(t, err)

		link, err := os.Readlink(hooksPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(newDir, HookDir), link)
	}
}
