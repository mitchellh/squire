package squire

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mitchellh/squire/internal/config"
)

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	// Build our config
	cfg, err := config.New(config.FromString(
		`sql_dir: "testdata/schema"`))
	require.NoError(err)

	// Build squire
	sq, err := New(WithConfig(cfg))
	require.NoError(err)

	// Get our container
	ctr, err := sq.Container()
	require.NoError(err)

	// Spin up the container
	require.NoError(ctr.Up(ctx))
	defer ctr.Down(ctx)

	// Connect
	db, err := ctr.Conn()
	require.NoError(err)
	defer db.Close()
	require.Eventually(func() bool {
		return db.Ping() == nil
	}, 5*time.Second, 10*time.Millisecond)

	// Test emptying
	require.NoError(recreateDB(ctx, ctr.Config()))
}
