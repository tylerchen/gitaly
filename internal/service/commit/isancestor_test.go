package commit

import (
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
)

const (
	scratchDir   = "testdata/scratch"
	testRepoRoot = "testdata/data"
	testRepo     = "group/test.git"
)

var serverSocketPath = path.Join(scratchDir, "gitaly.sock")

func TestMain(m *testing.M) {
	source := "https://gitlab.com/gitlab-org/gitlab-test.git"
	clonePath := path.Join(testRepoRoot, testRepo)
	if _, err := os.Stat(clonePath); err != nil {
		testCmd := exec.Command("git", "clone", "--bare", source, clonePath)
		testCmd.Stdout = os.Stdout
		testCmd.Stderr = os.Stderr

		if err := testCmd.Run(); err != nil {
			log.Printf("Test setup: failed to run %v", testCmd)
			os.Exit(-1)
		}
	}

	if err := os.MkdirAll(scratchDir, 0755); err != nil {
		log.Fatal(err)
	}

	os.Exit(func() int {
		os.Remove(serverSocketPath)
		server := runCommitServer(m)
		defer func() {
			server.Stop()
			os.Remove(serverSocketPath)
		}()

		return m.Run()
	}())
}

func TestCommitIsAncestorFailure(t *testing.T) {
	client := newCommitClient(t)
	repo := &pb.Repository{Path: path.Join(testRepoRoot, testRepo)}

	queries := []struct {
		Request   *pb.CommitIsAncestorRequest
		ErrorCode codes.Code
		ErrMsg    string
	}{
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: nil,
				AncestorId: "b83d6e391c22777fca1ed3012fce84f633d7fed0",
				ChildId:    "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab",
			},
			ErrorCode: codes.InvalidArgument,
			ErrMsg:    "Expected to throw invalid argument got: %s",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "",
				ChildId:    "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab",
			},
			ErrorCode: codes.InvalidArgument,
			ErrMsg:    "Expected to throw invalid argument got: %s",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "b83d6e391c22777fca1ed3012fce84f633d7fed0",
				ChildId:    "",
			},
			ErrorCode: codes.InvalidArgument,
			ErrMsg:    "Expected to throw invalid argument got: %s",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: &pb.Repository{Path: path.Join(testRepoRoot, testRepo, "2")},
				AncestorId: "b83d6e391c22777fca1ed3012fce84f633d7fed0",
				ChildId:    "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab",
			},
			ErrorCode: codes.Internal,
			ErrMsg:    "Expected to throw internal got: %s",
		},
	}

	for _, v := range queries {
		if _, err := client.CommitIsAncestor(context.Background(), v.Request); err == nil {
			t.Error("Expected to throw an error")
		} else if grpc.Code(err) != v.ErrorCode {
			t.Errorf(v.ErrMsg, err)
		}
	}
}

func TestCommitIsAncestorSuccess(t *testing.T) {
	client := newCommitClient(t)
	repo := &pb.Repository{Path: path.Join(testRepoRoot, testRepo)}

	queries := []struct {
		Request  *pb.CommitIsAncestorRequest
		Response bool
		ErrMsg   string
	}{
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab",
				ChildId:    "372ab6950519549b14d220271ee2322caa44d4eb",
			},
			Response: true,
			ErrMsg:   "Expected commit to be ancestor",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "b83d6e391c22777fca1ed3012fce84f633d7fed0",
				ChildId:    "38008cb17ce1466d8fec2dfa6f6ab8dcfe5cf49e",
			},
			Response: false,
			ErrMsg:   "Expected commit to not be ancestor",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "master",
				ChildId:    "gitaly-stuff",
			},
			Response: true,
			ErrMsg:   "Expected branch `master` to be ancestor of `gitaly-stuff`",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "gitaly-stuff",
				ChildId:    "master",
			},
			Response: false,
			ErrMsg:   "Expected branch `gitaly-stuff` not to be ancestor of `master`",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "refs/tags/v1.0.0",
				ChildId:    "refs/tags/v1.1.0",
			},
			Response: true,
			ErrMsg:   "Expected tag `v1.0.0` to be ancestor of `v1.1.0`",
		},
		{
			Request: &pb.CommitIsAncestorRequest{
				Repository: repo,
				AncestorId: "refs/tags/v1.1.0",
				ChildId:    "refs/tags/v1.0.0",
			},
			Response: false,
			ErrMsg:   "Expected branch `v1.1.0` not to be ancestor of `v1.0.0`",
		},
	}

	for _, v := range queries {
		c, err := client.CommitIsAncestor(context.Background(), v.Request)
		if err != nil {
			t.Fatalf("CommitIsAncestor threw error unexpectedly: %v", err)
		}

		response := c.GetValue()
		if response != v.Response {
			t.Errorf(v.ErrMsg)
		}
	}
}

func runCommitServer(m *testing.M) *grpc.Server {
	server := grpc.NewServer()
	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		log.Fatal(err)
	}

	pb.RegisterCommitServer(server, NewServer())
	reflection.Register(server)

	go server.Serve(listener)

	return server
}

func newCommitClient(t *testing.T) pb.CommitClient {
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

	return pb.NewCommitClient(conn)
}