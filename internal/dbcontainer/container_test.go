package dbcontainer

import (
	"context"
	"testing"
	"time"

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
	db, err := ctr.Conn()
	require.NoError(t, err)
	defer db.Close()
	require.Eventually(t, func() bool {
		return db.Ping() == nil
	}, 5*time.Second, 10*time.Millisecond)

	// Try cloning
	ctr2, err := ctr.Clone("dup")
	require.NoError(t, err)

	// Launch, ensure we come back down
	defer func() { require.NoError(t, ctr2.Down(ctx)) }()
	require.NoError(t, ctr2.Up(ctx))

	// Connect
	db2, err := ctr2.Conn()
	require.NoError(t, err)
	defer db2.Close()
	require.Eventually(t, func() bool {
		return db2.Ping() == nil
	}, 5*time.Second, 10*time.Millisecond)
}
