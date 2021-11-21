package squire

import (
	"bytes"
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
	db, err := ctr.Conn()
	require.NoError(err)
	defer db.Close()
	require.Eventually(func() bool {
		return db.Ping() == nil
	}, 5*time.Second, 10*time.Millisecond)

	// Generate schema
	var buf bytes.Buffer
	require.NoError(sq.Schema(&SchemaOptions{Output: &buf}))

	// Test deploy
	require.NoError(sq.Deploy(ctx, &DeployOptions{
		SQL:    &buf,
		Target: db,
	}))
}
