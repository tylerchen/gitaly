package repository

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/streamio"
)

func TestGetInfoAttributesExisting(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, repoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	infoPath := path.Join(repoPath, "info")
	os.MkdirAll(infoPath, 0755)

	buffSize := streamio.WriteBufferSize + 1
	data := bytes.Repeat([]byte("*.pbxproj binary\n"), buffSize)
	attrsPath := path.Join(infoPath, "attributes")
	err := ioutil.WriteFile(attrsPath, data, 0644)
	require.NoError(t, err)

	request := &pb.GetInfoAttributesRequest{Repository: testRepo}
	testCtx, cancelCtx := testhelper.Context()
	defer cancelCtx()

	stream, err := client.GetInfoAttributes(testCtx, request)
	require.NoError(t, err)

	receivedData, err := ioutil.ReadAll(streamio.NewReader(func() ([]byte, error) {
		response, err := stream.Recv()
		return response.GetAttributes(), err
	}))

	require.NoError(t, err)
	require.Equal(t, data, receivedData)
}

func TestGetInfoAttributesNonExisting(t *testing.T) {
	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	request := &pb.GetInfoAttributesRequest{Repository: testRepo}
	testCtx, cancelCtx := testhelper.Context()
	defer cancelCtx()

	response, err := client.GetInfoAttributes(testCtx, request)
	require.NoError(t, err)

	message, err := response.Recv()
	require.NoError(t, err)

	require.Empty(t, message.GetAttributes())
}