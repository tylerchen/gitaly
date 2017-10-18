package wiki

import (
	"net"
	"os"
	"path"
	"testing"
	"time"

	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	serverSocketPath = testhelper.GetTemporaryGitalySocketFileName()
	wikiRepo         *pb.Repository
	wikiRepoPath     string
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

var rubyServer *rubyserver.Server

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	testhelper.ConfigureTestStorage()
	storagePath := testhelper.GitlabTestStoragePath()
	wikiRepoPath = path.Join(storagePath, "wiki-test.git")

	testhelper.MustRunCommand(nil, nil, "git", "init", "--bare", wikiRepoPath)
	defer os.RemoveAll(wikiRepoPath)

	wikiRepo = &pb.Repository{
		StorageName:  "default",
		RelativePath: "wiki-test.git",
	}

	var err error
	testhelper.ConfigureRuby()
	rubyServer, err = rubyserver.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer rubyServer.Stop()

	return m.Run()
}

func runWikiServiceServer(t *testing.T) *grpc.Server {
	os.Remove(serverSocketPath)
	grpcServer := testhelper.NewTestGrpcServer(t, nil, nil)

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal(err)
	}

	pb.RegisterWikiServiceServer(grpcServer, &server{rubyServer})
	reflection.Register(grpcServer)

	go grpcServer.Serve(listener)

	return grpcServer
}

func newWikiClient(t *testing.T) (pb.WikiServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, _ time.Duration) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return pb.NewWikiServiceClient(conn), conn
}