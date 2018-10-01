package operations

import (
	"os"
	"os/exec"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"golang.org/x/net/context"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
)

func TestSuccessfulUserCreateBranchRequest(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	startPoint := "c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"
	startPointCommit, err := log.GetCommit(ctx, testRepo, startPoint)
	require.NoError(t, err)

	testCases := []struct {
		desc           string
		branchName     string
		startPoint     string
		expectedBranch *pb.Branch
	}{
		{
			desc:       "valid branch",
			branchName: "new-branch",
			startPoint: startPoint,
			expectedBranch: &pb.Branch{
				Name:         []byte("new-branch"),
				TargetCommit: startPointCommit,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			branchName := testCase.branchName
			request := &pb.UserCreateBranchRequest{
				Repository: testRepo,
				BranchName: []byte(branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       user,
			}

			ctx, cancel := testhelper.Context()
			defer cancel()

			response, err := client.UserCreateBranch(ctx, request)
			if testCase.expectedBranch != nil {
				defer exec.Command("git", "-C", testRepoPath, "branch", "-D", branchName).Run()
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expectedBranch, response.Branch)
			require.Empty(t, response.PreReceiveError)

			branches := testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch")
			require.Contains(t, string(branches), branchName)
		})
	}
}

func TestSuccessfulGitHooksForUserCreateBranchRequest(t *testing.T) {
	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	branchName := "new-branch"
	request := &pb.UserCreateBranchRequest{
		Repository: testRepo,
		BranchName: []byte(branchName),
		StartPoint: []byte("c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"),
		User:       user,
	}

	for _, hookName := range GitlabHooks {
		t.Run(hookName, func(t *testing.T) {
			defer exec.Command("git", "-C", testRepoPath, "branch", "-D", branchName).Run()

			hookPath, hookOutputTempPath := WriteEnvToHook(t, testRepoPath, hookName)
			defer os.Remove(hookPath)
			defer os.Remove(hookOutputTempPath)

			ctx, cancel := testhelper.Context()
			defer cancel()

			response, err := client.UserCreateBranch(ctx, request)
			require.NoError(t, err)
			require.Empty(t, response.PreReceiveError)

			output := string(testhelper.MustReadFile(t, hookOutputTempPath))
			require.Contains(t, output, "GL_ID="+user.GlId)
			require.Contains(t, output, "GL_USERNAME="+user.GlUsername)
		})
	}
}

func TestFailedUserCreateBranchDueToHooks(t *testing.T) {
	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	request := &pb.UserCreateBranchRequest{
		Repository: testRepo,
		BranchName: []byte("new-branch"),
		StartPoint: []byte("c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"),
		User:       user,
	}
	// Write a hook that will fail with the environment as the error message
	// so we can check that string for our env variables.
	hookContent := []byte("#!/bin/sh\nprintenv | paste -sd ' ' -\nexit 1")

	for _, hookName := range gitlabPreHooks {
		cleanFn, err := OverrideHooks(testRepoPath, hookName, hookContent)
		require.NoError(t, err)
		defer cleanFn()

		ctx, cancel := testhelper.Context()
		defer cancel()

		response, err := client.UserCreateBranch(ctx, request)
		require.Nil(t, err)
		require.Contains(t, response.PreReceiveError, "GL_ID="+user.GlId)
		require.Contains(t, response.PreReceiveError, "GL_USERNAME="+user.GlUsername)
		require.Contains(t, response.PreReceiveError, "GL_REPOSITORY="+testRepo.GlRepository)
		require.Contains(t, response.PreReceiveError, "GL_PROTOCOL=web")
		require.Contains(t, response.PreReceiveError, "PWD="+testRepoPath)
	}
}

func TestFailedUserCreateBranchRequest(t *testing.T) {
	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	testCases := []struct {
		desc       string
		branchName string
		startPoint string
		user       *pb.User
		code       codes.Code
	}{
		{
			desc:       "empty start_point",
			branchName: "shiny-new-branch",
			startPoint: "",
			user:       user,
			code:       codes.InvalidArgument,
		},
		{
			desc:       "empty user",
			branchName: "shiny-new-branch",
			startPoint: "master",
			user:       nil,
			code:       codes.InvalidArgument,
		},
		{
			desc:       "non-existing starting point",
			branchName: "new-branch",
			startPoint: "i-dont-exist",
			user:       user,
			code:       codes.FailedPrecondition,
		},

		{
			desc:       "branch exists",
			branchName: "master",
			startPoint: "master",
			user:       user,
			code:       codes.FailedPrecondition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			request := &pb.UserCreateBranchRequest{
				Repository: testRepo,
				BranchName: []byte(testCase.branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       testCase.user,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			_, err := client.UserCreateBranch(ctx, request)
			testhelper.RequireGrpcError(t, err, testCase.code)
		})
	}
}

func TestSuccessfulUserDeleteBranchRequest(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	branchNameInput := "to-be-deleted-soon-branch"

	defer exec.Command("git", "-C", testRepoPath, "branch", "-d", branchNameInput).Run()

	user := &pb.User{
		Name:  []byte("Alejandro Rodríguez"),
		Email: []byte("alejandro@gitlab.com"),
		GlId:  "user-123",
	}

	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", branchNameInput)

	request := &pb.UserDeleteBranchRequest{
		Repository: testRepo,
		BranchName: []byte(branchNameInput),
		User:       user,
	}

	_, err := client.UserDeleteBranch(ctx, request)
	require.NoError(t, err)

	branches := testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch")
	require.NotContains(t, string(branches), branchNameInput, "branch name still exists in branches list")
}

func TestSuccessfulGitHooksForUserDeleteBranchRequest(t *testing.T) {
	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	branchNameInput := "to-be-deleted-soon-branch"
	defer exec.Command("git", "-C", testRepoPath, "branch", "-d", branchNameInput).Run()

	user := &pb.User{
		Name:       []byte("Alejandro Rodríguez"),
		Email:      []byte("alejandro@gitlab.com"),
		GlId:       "user-123",
		GlUsername: "johndoe",
	}

	request := &pb.UserDeleteBranchRequest{
		Repository: testRepo,
		BranchName: []byte(branchNameInput),
		User:       user,
	}

	for _, hookName := range GitlabHooks {
		t.Run(hookName, func(t *testing.T) {
			testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", branchNameInput)

			hookPath, hookOutputTempPath := WriteEnvToHook(t, testRepoPath, hookName)
			defer os.Remove(hookPath)
			defer os.Remove(hookOutputTempPath)

			ctx, cancel := testhelper.Context()
			defer cancel()

			_, err := client.UserDeleteBranch(ctx, request)
			require.NoError(t, err)

			output := testhelper.MustReadFile(t, hookOutputTempPath)
			require.Contains(t, string(output), "GL_ID=user-123")
			require.Contains(t, string(output), "GL_USERNAME=johndoe")
		})
	}
}

func TestFailedUserDeleteBranchDueToValidation(t *testing.T) {
	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	user := &pb.User{
		Name:  []byte("Alejandro Rodríguez"),
		Email: []byte("alejandro@gitlab.com"),
		GlId:  "user-123",
	}

	testCases := []struct {
		desc    string
		request *pb.UserDeleteBranchRequest
		code    codes.Code
	}{
		{
			desc: "empty user",
			request: &pb.UserDeleteBranchRequest{
				Repository: testRepo,
				BranchName: []byte("does-matter-the-name-if-user-is-empty"),
			},
			code: codes.InvalidArgument,
		},
		{
			desc: "empty branch name",
			request: &pb.UserDeleteBranchRequest{
				Repository: testRepo,
				User:       user,
			},
			code: codes.InvalidArgument,
		},
		{
			desc: "non-existent branch name",
			request: &pb.UserDeleteBranchRequest{
				Repository: testRepo,
				User:       user,
				BranchName: []byte("i-do-not-exist"),
			},
			code: codes.FailedPrecondition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			ctx, cancel := testhelper.Context()
			defer cancel()

			_, err := client.UserDeleteBranch(ctx, testCase.request)
			testhelper.RequireGrpcError(t, err, testCase.code)
		})
	}
}

func TestFailedUserDeleteBranchDueToHooks(t *testing.T) {
	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	server, serverSocketPath := runOperationServiceServer(t)
	defer server.Stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	branchNameInput := "to-be-deleted-soon-branch"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", branchNameInput)
	defer exec.Command("git", "-C", testRepoPath, "branch", "-d", branchNameInput).Run()

	user := &pb.User{
		Name:  []byte("Alejandro Rodríguez"),
		Email: []byte("alejandro@gitlab.com"),
		GlId:  "user-123",
	}

	request := &pb.UserDeleteBranchRequest{
		Repository: testRepo,
		BranchName: []byte(branchNameInput),
		User:       user,
	}

	hookContent := []byte("#!/bin/sh\necho GL_ID=$GL_ID\nexit 1")

	for _, hookName := range gitlabPreHooks {
		t.Run(hookName, func(t *testing.T) {
			cleanFn, err := OverrideHooks(testRepoPath, hookName, hookContent)
			require.NoError(t, err)
			defer cleanFn()

			ctx, cancel := testhelper.Context()
			defer cancel()

			response, err := client.UserDeleteBranch(ctx, request)
			require.Nil(t, err)
			require.Contains(t, response.PreReceiveError, "GL_ID="+user.GlId)

			branches := testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch")
			require.Contains(t, string(branches), branchNameInput, "branch name does not exist in branches list")
		})
	}
}
