package squire

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mitchellh/squire/internal/config"
)

func TestDeploy(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	// Build our config
	cfg, err := config.New(config.FromString(
		`sql_dir: "testdata/deploy"`))
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
	db, err := ctr.Conn(ctx)
	require.NoError(err)
	defer db.Close()
	require.NoError(db.Ping())

	// Generate schema
	var buf bytes.Buffer
	require.NoError(sq.Schema(&SchemaOptions{Output: &buf}))

	// Test deploy
	require.NoError(sq.Deploy(ctx, &DeployOptions{
		SQL:    &buf,
		Target: db,
	}))

	// Test a clean reset
	require.NoError(sq.Reset(ctx, &ResetOptions{
		Container: ctr,
	}))
}
