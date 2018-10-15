package smarthttp

import (
	"io"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/streamio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	deepenCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gitaly_smarthttp_deepen_count",
			Help: "Number of git-upload-pack requests processed that contained a 'deepen' message",
		},
	)
)

func init() {
	prometheus.MustRegister(deepenCount)
}

func (s *server) PostUploadPack(stream gitalypb.SmartHTTPService_PostUploadPackServer) error {
	ctx := stream.Context()

	req, err := stream.Recv() // First request contains Repository only
	if err != nil {
		return err
	}

	if err := validateUploadPackRequest(req); err != nil {
		return err
	}

	stdinReader := streamio.NewReader(func() ([]byte, error) {
		resp, err := stream.Recv()
		return resp.GetData(), err
	})
	pr, pw := io.Pipe()
	defer pw.Close()
	stdin := io.TeeReader(stdinReader, pw)
	deepenCh := make(chan bool, 1)
	go func() {
		deepenCh <- scanDeepen(pr)
	}()

	stdout := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&gitalypb.PostUploadPackResponse{Data: p})
	})

	env := git.AddGitProtocolEnv(ctx, req, command.GitEnv)

	repoPath, err := helper.GetRepoPath(req.Repository)
	if err != nil {
		return err
	}

	args := []string{}
	for _, params := range req.GitConfigOptions {
		args = append(args, "-c", params)
	}

	args = append(args, "upload-pack", "--stateless-rpc", repoPath)

	cmd, err := git.BareCommand(ctx, stdin, stdout, nil, env, args...)

	if err != nil {
		return status.Errorf(codes.Unavailable, "PostUploadPack: cmd: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		pw.Close() // ensure scanDeepen returns
		if _, ok := command.ExitStatus(err); ok && <-deepenCh {
			// We have seen a 'deepen' message in the request. It is expected that
			// git-upload-pack has a non-zero exit status: don't treat this as an
			// error.
			deepenCount.Inc()
			return nil
		}
		return status.Errorf(codes.Unavailable, "PostUploadPack: %v", err)
	}

	return nil
}

func validateUploadPackRequest(req *gitalypb.PostUploadPackRequest) error {
	if req.Data != nil {
		return status.Errorf(codes.InvalidArgument, "PostUploadPack: non-empty Data")
	}

	return nil
}
