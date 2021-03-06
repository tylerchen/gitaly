package repository

import (
	"bytes"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"google.golang.org/grpc/codes"
)

func TestWriteRefSuccessful(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	testCases := []struct {
		desc string
		req  *gitalypb.WriteRefRequest
	}{
		{
			desc: "rugged update refs/heads/master",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/heads/master"),
				Revision:   []byte("4a24d82dbca5c11c61556f3b35ca472b7463187e"),
				Shell:      false,
			},
		},
		{
			desc: "rugged update refs/keep-around/4a24d82dbca5c11c61556f3b35ca472b7463187e",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/keep-around/4a24d82dbca5c11c61556f3b35ca472b7463187e"),
				Revision:   []byte("4a24d82dbca5c11c61556f3b35ca472b7463187e"),
				Shell:      false,
			},
		},
		{
			desc: "rugged update HEAD to refs/heads/master",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("HEAD"),
				Revision:   []byte("refs/heads/master"),
				Shell:      false,
			},
		},
		{
			desc: "shell update refs/heads/master",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/heads/master"),
				Revision:   []byte("b83d6e391c22777fca1ed3012fce84f633d7fed0"),
				Shell:      true,
			},
		},
		{
			desc: "shell update refs/heads/master w/ validation",
			req: &gitalypb.WriteRefRequest{
				Repository:  testRepo,
				Ref:         []byte("refs/heads/master"),
				Revision:    []byte("498214de67004b1da3d820901307bed2a68a8ef6"),
				OldRevision: []byte("b83d6e391c22777fca1ed3012fce84f633d7fed0"),
				Shell:       true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := testhelper.Context()
			defer cancel()
			_, err := client.WriteRef(ctx, tc.req)

			require.NoError(t, err)

			if bytes.Equal(tc.req.Ref, []byte("HEAD")) {
				content := testhelper.MustReadFile(t, path.Join(testRepoPath, "HEAD"))

				refRevision := bytes.Join([][]byte{[]byte("ref: "), tc.req.Revision, []byte("\n")}, nil)

				require.EqualValues(t, content, refRevision)
				return
			}
			rev := testhelper.MustRunCommand(t, nil, "git", "--git-dir", testRepoPath, "log", "--pretty=%H", "-1", string(tc.req.Ref))

			rev = bytes.Replace(rev, []byte("\n"), nil, 1)

			require.Equal(t, string(tc.req.Revision), string(rev))
		})
	}
}

func TestWriteRefValidationError(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	testCases := []struct {
		desc string
		req  *gitalypb.WriteRefRequest
	}{
		{
			desc: "empty revision",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/heads/master"),
			},
		},
		{
			desc: "empty ref name",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Revision:   []byte("498214de67004b1da3d820901307bed2a68a8ef6"),
			},
		},
		{
			desc: "non-prefixed ref name for shell",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("master"),
				Revision:   []byte("498214de67004b1da3d820901307bed2a68a8ef6"),
				Shell:      true,
			},
		},
		{
			desc: "revision contains \\x00",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/heads/master"),
				Revision:   []byte("012301230123\x001243"),
			},
		},
		{
			desc: "ref contains \\x00",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/head\x00s/master\x00"),
				Revision:   []byte("0123012301231243"),
			},
		},
		{
			desc: "ref contains whitespace",
			req: &gitalypb.WriteRefRequest{
				Repository: testRepo,
				Ref:        []byte("refs/heads /master"),
				Revision:   []byte("0123012301231243"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := testhelper.Context()
			defer cancel()
			_, err := client.WriteRef(ctx, tc.req)

			testhelper.RequireGrpcError(t, err, codes.InvalidArgument)
		})
	}
}
