package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configFileReader(content string) io.Reader {
	return bytes.NewReader([]byte(content))
}

func TestLoadClearPrevConfig(t *testing.T) {
	Config = config{SocketPath: "/tmp"}
	err := Load(&bytes.Buffer{})
	assert.NoError(t, err)

	assert.Empty(t, Config.SocketPath)
}

func TestLoadBrokenConfig(t *testing.T) {
	tmpFile := configFileReader(`path = "/tmp"\nname="foo"`)
	err := Load(tmpFile)
	assert.Error(t, err)

	assert.Equal(t, config{}, Config)
}

func TestLoadEmptyConfig(t *testing.T) {
	tmpFile := configFileReader(``)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, config{}, Config)
}

func TestLoadStorage(t *testing.T) {
	tmpFile := configFileReader(`[[storage]]
name = "default"
path = "/tmp"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	if assert.Equal(t, 1, len(Config.Storages), "Expected one (1) storage") {
		assert.Equal(t, config{
			Storages: []Storage{
				{Name: "default", Path: "/tmp"},
			},
		}, Config)
	}
}

func TestLoadMultiStorage(t *testing.T) {
	tmpFile := configFileReader(`[[storage]]
name="default"
path="/tmp/repos1"

[[storage]]
name="other"
path="/tmp/repos2"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	if assert.Equal(t, 2, len(Config.Storages), "Expected one (1) storage") {
		assert.Equal(t, config{
			Storages: []Storage{
				{Name: "default", Path: "/tmp/repos1"},
				{Name: "other", Path: "/tmp/repos2"},
			},
		}, Config)
	}
}

func TestLoadPrometheus(t *testing.T) {
	tmpFile := configFileReader(`prometheus_listen_addr=":9236"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, ":9236", Config.PrometheusListenAddr)
}

func TestLoadSocketPath(t *testing.T) {
	tmpFile := configFileReader(`socket_path="/tmp/gitaly.sock"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, "/tmp/gitaly.sock", Config.SocketPath)
}

func TestLoadListenAddr(t *testing.T) {
	tmpFile := configFileReader(`listen_addr=":8080"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, ":8080", Config.ListenAddr)
}

func tempEnv(key, value string) func() {
	temp := os.Getenv(key)
	os.Setenv(key, value)

	return func() {
		os.Setenv(key, temp)
	}
}

func TestLoadOverrideEnvironment(t *testing.T) {
	// Test that this works since we still want this to work
	tempEnv1 := tempEnv("GITALY_SOCKET_PATH", "/tmp/gitaly2.sock")
	defer tempEnv1()
	tempEnv2 := tempEnv("GITALY_LISTEN_ADDR", ":8081")
	defer tempEnv2()
	tempEnv3 := tempEnv("GITALY_PROMETHEUS_LISTEN_ADDR", ":9237")
	defer tempEnv3()

	tmpFile := configFileReader(`socket_path = "/tmp/gitaly.sock"
listen_addr = ":8080"
prometheus_listen_addr = ":9236"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, ":9237", Config.PrometheusListenAddr)
	assert.Equal(t, "/tmp/gitaly2.sock", Config.SocketPath)
	assert.Equal(t, ":8081", Config.ListenAddr)
}

func TestLoadOnlyEnvironment(t *testing.T) {
	// Test that this works since we still want this to work
	os.Setenv("GITALY_SOCKET_PATH", "/tmp/gitaly2.sock")
	os.Setenv("GITALY_LISTEN_ADDR", ":8081")
	os.Setenv("GITALY_PROMETHEUS_LISTEN_ADDR", ":9237")

	err := Load(&bytes.Buffer{})
	assert.NoError(t, err)

	assert.Equal(t, ":9237", Config.PrometheusListenAddr)
	assert.Equal(t, "/tmp/gitaly2.sock", Config.SocketPath)
	assert.Equal(t, ":8081", Config.ListenAddr)
}

func TestValidateStorages(t *testing.T) {
	defer func(oldStorages []Storage) {
		Config.Storages = oldStorages
	}(Config.Storages)

	tmpDir := os.TempDir()
	workDir, err := os.Getwd()
	require.NoError(t, err)
	invalidDir := path.Join(tmpDir, t.Name())

	testCases := []struct {
		desc     string
		storages []Storage
		invalid  bool
	}{
		{
			desc: "just 1 storage",
			storages: []Storage{
				{Name: "default", Path: tmpDir},
			},
		},
		{
			desc: "multiple storages",
			storages: []Storage{
				{Name: "default", Path: tmpDir},
				{Name: "other", Path: workDir},
			},
		},
		{
			desc: "multiple storages pointing to same directory",
			storages: []Storage{
				{Name: "default", Path: tmpDir},
				{Name: "other", Path: tmpDir},
				{Name: "third", Path: tmpDir},
			},
		},
		{
			desc: "nested paths 1",
			storages: []Storage{
				{Name: "default", Path: "/home/git/repositories"},
				{Name: "other", Path: "/home/git/repositories"},
				{Name: "third", Path: "/home/git/repositories/third"},
			},
			invalid: true,
		},
		{
			desc: "nested paths 2",
			storages: []Storage{
				{Name: "default", Path: "/home/git/repositories/default"},
				{Name: "other", Path: "/home/git/repositories"},
				{Name: "third", Path: "/home/git/repositories"},
			},
			invalid: true,
		},
		{
			desc: "duplicate definition",
			storages: []Storage{
				{Name: "default", Path: tmpDir},
				{Name: "default", Path: tmpDir},
			},
			invalid: true,
		},
		{
			desc: "re-definition",
			storages: []Storage{
				{Name: "default", Path: tmpDir},
				{Name: "default", Path: invalidDir},
			},
			invalid: true,
		},
		{
			desc: "empty name",
			storages: []Storage{
				{Name: "", Path: tmpDir},
			},
			invalid: true,
		},
		{
			desc: "empty path",
			storages: []Storage{
				{Name: "default", Path: ""},
			},
			invalid: true,
		},
		{
			storages: []Storage{
				{Name: "default", Path: tmpDir},
				{Name: "nope", Path: invalidDir},
			},
			invalid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			Config.Storages = tc.storages
			err := validateStorages()
			if tc.invalid {
				assert.Error(t, err, "%+v", tc.storages)
				return
			}

			assert.NoError(t, err, "%+v", tc.storages)
		})
	}
}

func TestStoragePath(t *testing.T) {
	defer func(oldStorages []Storage) {
		Config.Storages = oldStorages
	}(Config.Storages)

	Config.Storages = []Storage{
		{Name: "default", Path: "/home/git/repositories1"},
		{Name: "other", Path: "/home/git/repositories2"},
		{Name: "third", Path: "/home/git/repositories3"},
	}

	testCases := []struct {
		in, out string
		ok      bool
	}{
		{in: "default", out: "/home/git/repositories1", ok: true},
		{in: "third", out: "/home/git/repositories3", ok: true},
		{in: "", ok: false},
		{in: "foobar", ok: false},
	}

	for _, tc := range testCases {
		out, ok := StoragePath(tc.in)
		if !assert.Equal(t, tc.ok, ok, "%+v", tc) {
			continue
		}
		assert.Equal(t, tc.out, out, "%+v", tc)
	}
}

func TestLoadGit(t *testing.T) {
	tmpFile := configFileReader(`[git]
bin_path = "/my/git/path"`)

	err := Load(tmpFile)
	assert.NoError(t, err)

	assert.Equal(t, "/my/git/path", Config.Git.BinPath)
}

func TestSetGitPath(t *testing.T) {
	defer func(oldGitSettings Git) {
		Config.Git = oldGitSettings
	}(Config.Git)

	resolvedGitPath, err := exec.LookPath("git")

	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		desc       string
		gitBinPath string
		expected   string
	}{
		{
			desc:       "With a Git Path set through the settings",
			gitBinPath: "/path/to/myGit",
			expected:   "/path/to/myGit",
		},
		{
			desc:       "When a git path hasn't been set",
			gitBinPath: "",
			expected:   resolvedGitPath,
		},
	}

	for _, tc := range testCases {
		Config.Git.BinPath = tc.gitBinPath

		SetGitPath()

		assert.Equal(t, tc.expected, Config.Git.BinPath, tc.desc)
	}
}

func TestValidateShellPath(t *testing.T) {
	defer func(oldShellSettings GitlabShell) {
		Config.GitlabShell = oldShellSettings
	}(Config.GitlabShell)

	tmpDir, err := ioutil.TempDir("", "gitaly-tests-")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(path.Join(tmpDir, "bin"), 0755))
	tmpFile := path.Join(tmpDir, "my-file")
	defer os.RemoveAll(tmpDir)
	fp, err := os.Create(tmpFile)
	require.NoError(t, err)
	require.NoError(t, fp.Close())

	testCases := []struct {
		desc      string
		path      string
		shouldErr bool
	}{
		{
			desc:      "When no Shell Path set",
			path:      "",
			shouldErr: true,
		},
		{
			desc:      "When Shell Path set to non-existing path",
			path:      "/non/existing/path",
			shouldErr: true,
		},
		{
			desc:      "When Shell Path set to non-dir path",
			path:      tmpFile,
			shouldErr: true,
		},
		{
			desc:      "When Shell Path set to a valid directory",
			path:      tmpDir,
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Log(tc.desc)
		Config.GitlabShell.Dir = tc.path
		err = validateShell()
		if tc.shouldErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

	}
}

func TestConfigureRuby(t *testing.T) {
	defer func(oldRuby Ruby) {
		Config.Ruby = oldRuby
	}(Config.Ruby)

	tmpDir, err := ioutil.TempDir("", "gitaly-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tmpFile := path.Join(tmpDir, "file")
	require.NoError(t, ioutil.WriteFile(tmpFile, nil, 0644))

	testCases := []struct {
		dir  string
		ok   bool
		desc string
	}{
		{dir: "", desc: "empty"},
		{dir: "/does/not/exist", desc: "does not exist"},
		{dir: tmpFile, desc: "exists but is not a directory"},
		{dir: tmpDir, ok: true, desc: "ok"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			Config.Ruby = Ruby{Dir: tc.dir}

			err := ConfigureRuby()
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestConfigureRubyNumWorkers(t *testing.T) {
	defer func(oldRuby Ruby) {
		Config.Ruby = oldRuby
	}(Config.Ruby)

	testCases := []struct {
		in, out int
	}{
		{in: -1, out: 2},
		{in: 0, out: 2},
		{in: 1, out: 2},
		{in: 2, out: 2},
		{in: 3, out: 3},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			Config.Ruby = Ruby{Dir: "/", NumWorkers: tc.in}
			require.NoError(t, ConfigureRuby())
			require.Equal(t, tc.out, Config.Ruby.NumWorkers)
		})
	}
}

func TestValidateListeners(t *testing.T) {
	defer func(cfg config) {
		Config = cfg
	}(Config)

	testCases := []struct {
		desc string
		config
		ok bool
	}{
		{desc: "empty"},
		{desc: "socket only", config: config{SocketPath: "/foo/bar"}, ok: true},
		{desc: "tcp only", config: config{ListenAddr: "a.b.c.d:1234"}, ok: true},
		{desc: "both socket and tcp", config: config{SocketPath: "/foo/bar", ListenAddr: "a.b.c.d:1234"}, ok: true},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			Config = tc.config
			err := validateListeners()
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
