package dbcontainer

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	//"github.com/davecgh/go-spew/spew"

	"github.com/mitchellh/squire/internal/dbcompose"
)

// This is a really big test because we don't want to spin up a ton of
// containers cause they're slow.
func TestUpDown(t *testing.T) {
	ctx := context.Background()

	// init config
	cfg, err := dbcompose.New(dbcompose.WithPath("testdata/compose-v2.yml"))
	require.NoError(t, err)

	// Init
	ctr, err := New(WithCompose(cfg))
	require.NoError(t, err)

	// Launch, ensure we come back down
	defer func() {
		require.NoError(t, ctr.Down(ctx))
	}()
	require.NoError(t, ctr.Up(ctx))

	// Connect
	db, err := sql.Open("postgres", ctr.ConnURI())
	require.NoError(t, err)
	defer db.Close()

	// Try cloning
	ctr2, err := ctr.Clone("dup")
	require.NoError(t, err)

	// Launch, ensure we come back down
	defer func() { require.NoError(t, ctr2.Down(ctx)) }()
	require.NoError(t, ctr2.Up(ctx))

	// Connect
	db2, err := sql.Open("postgres", ctr2.ConnURI())
	require.NoError(t, err)
	defer db2.Close()
}
