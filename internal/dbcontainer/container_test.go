package dbcontainer

import (
	"context"
	"testing"

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

	// Test status, should be not created
	st, err := ctr.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, NotCreated, st.State)

	// Launch, ensure we come back down
	defer func() {
		require.NoError(t, ctr.Down(ctx))
	}()
	require.NoError(t, ctr.Up(ctx))

	// Connect
	db, err := ctr.Conn(ctx)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.Ping())

	// Should be created
	st, err = ctr.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, Running, st.State)

	// Try cloning
	ctr2, err := ctr.Clone("dup")
	require.NoError(t, err)

	// Launch, ensure we come back down
	defer func() { require.NoError(t, ctr2.Down(ctx)) }()
	require.NoError(t, ctr2.Up(ctx))

	// Connect
	db2, err := ctr2.Conn(ctx)
	require.NoError(t, err)
	defer db2.Close()
	require.NoError(t, db2.Ping())

	// Bring first container down
	require.NoError(t, ctr.Down(ctx))

	// Should be not created
	st, err = ctr.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, NotCreated, st.State)
}
