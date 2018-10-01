package operations

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"gitlab.com/gitlab-org/gitaly/internal/config"
	"gitlab.com/gitlab-org/gitaly/internal/git/hooks"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	gitlabPreHooks  = []string{"pre-receive", "update"}
	gitlabPostHooks = []string{"post-receive"}
	GitlabPreHooks  = gitlabPreHooks
	GitlabHooks     []string
	RubyServer      *rubyserver.Server
	user            = &pb.User{
		Name:       []byte("Jane Doe"),
		Email:      []byte("janedoe@gitlab.com"),
		GlId:       "user-123",
		GlUsername: "janedoe",
	}
)

func init() {
	copy(GitlabHooks, gitlabPreHooks)
	GitlabHooks = append(GitlabHooks, gitlabPostHooks...)
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	var err error

	testhelper.ConfigureGitalySSH()

	testhelper.ConfigureRuby()
	RubyServer, err = rubyserver.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer RubyServer.Stop()

	return m.Run()
}

func runOperationServiceServer(t *testing.T) (*grpc.Server, string) {
	grpcServer := testhelper.NewTestGrpcServer(t, nil, nil)
	serverSocketPath := testhelper.GetTemporaryGitalySocketFileName()

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal(err)
	}

	pb.RegisterOperationServiceServer(grpcServer, &server{RubyServer})
	reflection.Register(grpcServer)

	go grpcServer.Serve(listener)

	return grpcServer, serverSocketPath
}

func newOperationClient(t *testing.T, serverSocketPath string) (pb.OperationServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return pb.NewOperationServiceClient(conn), conn
}

var NewOperationClient = newOperationClient

// Caller is responsible for cleanup of the repository
func WriteEnvToHook(t *testing.T, repoPath, hookName string) (string, string) {
	hookOutputTemp, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	require.NoError(t, hookOutputTemp.Close())

	defer os.Remove(hookOutputTemp.Name())

	hookContent := fmt.Sprintf("#!/bin/sh\n/usr/bin/env > %s\n", hookOutputTemp.Name())

	_, err = OverrideHooks(repoPath, hookName, []byte(hookContent))
	require.NoError(t, err)

	return path.Join(repoPath, "hooks", hookName), hookOutputTemp.Name()
}

// OverrrideHooks sets the hooks location to its repoPath, with a custom
// directory, so the repository from which we seed isn't touched and the test
// doesn't require write permission to that location
func OverrideHooks(repoPath, name string, content []byte) (func(), error) {
	gitlabShellDir := config.Config.GitlabShell.Dir
	config.Config.GitlabShell.Dir = path.Join(repoPath, "tmphook")

	if err := os.MkdirAll(hooks.Path(), 0755); err != nil {
		return nil, err
	}

	fullPath := path.Join(hooks.Path(), name)
	err := ioutil.WriteFile(fullPath, content, 0777)

	cleanFn := func() {
		config.Config.GitlabShell.Dir = gitlabShellDir
		os.RemoveAll(fullPath)
	}

	return cleanFn, err
}
