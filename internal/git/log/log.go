package log

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
)

// Parser holds necessary state for parsing a git log stream
type Parser struct {
	scanner       *bufio.Scanner
	currentCommit *gitalypb.GitCommit
	err           error
	c             *catfile.Batch
}

// NewLogParser returns a new Parser
func NewLogParser(ctx context.Context, repo *gitalypb.Repository, src io.Reader) (*Parser, error) {
	c, err := catfile.New(ctx, repo)
	if err != nil {
		return nil, err
	}

	parser := &Parser{
		scanner: bufio.NewScanner(src),
		c:       c,
	}

	return parser, nil
}

// Parse parses a single git log line. It returns true if successful, false if it finished
// parsing all logs or when it encounters an error, in which case use Parser.Err()
// to get the error.
func (parser *Parser) Parse() bool {
	if !parser.scanner.Scan() || parser.err != nil {
		return false
	}

	commitID := parser.scanner.Text()

	commit, err := GetCommitCatfile(parser.c, commitID)
	if err != nil {
		parser.err = err
		return false
	}

	if commit == nil {
		parser.err = fmt.Errorf("could not retrieve commit %q", commitID)
		return false
	}

	parser.currentCommit = commit
	return true
}

// Commit returns a successfully parsed git log line. It should be called only when Parser.Parse()
// returns true.
func (parser *Parser) Commit() *gitalypb.GitCommit {
	return parser.currentCommit
}

// Err returns the error encountered (if any) when parsing the diff stream. It should be called only when Parser.Parse()
// returns false.
func (parser *Parser) Err() error {
	if parser.err == nil {
		parser.err = parser.scanner.Err()
	}

	return parser.err
}
