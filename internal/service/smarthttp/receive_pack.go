package smarthttp

import (
	"fmt"
	"os/exec"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/hooks"
	"gitlab.com/gitlab-org/gitaly/internal/helper"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/streamio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) PostReceivePack(stream pb.SmartHTTPService_PostReceivePackServer) error {
	req, err := stream.Recv() // First request contains only Repository and GlId
	if err != nil {
		return err
	}

	grpc_logrus.Extract(stream.Context()).WithFields(log.Fields{
		"GlID":             req.GlId,
		"GlRepository":     req.GlRepository,
		"GlUsername":       req.GlUsername,
		"GitConfigOptions": req.GitConfigOptions,
	}).Debug("PostReceivePack")

	if err := validateReceivePackRequest(req); err != nil {
		return err
	}

	stdin := streamio.NewReader(func() ([]byte, error) {
		resp, err := stream.Recv()
		return resp.GetData(), err
	})
	stdout := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&pb.PostReceivePackResponse{Data: p})
	})
	env := []string{
		fmt.Sprintf("GL_ID=%s", req.GlId),
		"GL_PROTOCOL=http",
	}
	if req.GlRepository != "" {
		env = append(env, fmt.Sprintf("GL_REPOSITORY=%s", req.GlRepository))
	}
	if req.GlUsername != "" {
		env = append(env, fmt.Sprintf("GL_USERNAME=%s", req.GlUsername))
	}

	repoPath, err := helper.GetRepoPath(req.Repository)
	if err != nil {
		return err
	}

	if err := hooks.SetGitLabHooks(repoPath); err != nil {
		return err
	}

	gitOptions := git.BuildGitOptions(req.GitConfigOptions, "receive-pack", "--stateless-rpc", repoPath)
	osCommand := exec.Command(command.GitPath(), gitOptions...)
	cmd, err := command.New(stream.Context(), osCommand, stdin, stdout, nil, env...)

	if err != nil {
		return status.Errorf(codes.Unavailable, "PostReceivePack: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return status.Errorf(codes.Unavailable, "PostReceivePack: %v", err)
	}

	return nil
}

func validateReceivePackRequest(req *pb.PostReceivePackRequest) error {
	if req.GlId == "" {
		return status.Errorf(codes.InvalidArgument, "PostReceivePack: empty GlId")
	}
	if req.Data != nil {
		return status.Errorf(codes.InvalidArgument, "PostReceivePack: non-empty Data")
	}

	return nil
}
