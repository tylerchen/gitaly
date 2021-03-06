package commit

import (
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/service/ref"
)

type server struct {
	*rubyserver.Server
}

var (
	defaultBranchName = ref.DefaultBranchName
)

// NewServer creates a new instance of a grpc CommitServiceServer
func NewServer(rs *rubyserver.Server) gitalypb.CommitServiceServer {
	return &server{rs}
}
