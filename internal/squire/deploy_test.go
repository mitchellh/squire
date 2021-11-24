package squire

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
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

	// Try invalid SQL so that we can verify our error message
	// has useful contents.
	err = sq.Deploy(ctx, &DeployOptions{
		SQL: strings.NewReader(strings.TrimSpace(`
CREATE TABLE accounts (
  id         SERIAL PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
);
`)),
		Target: db,
	})
	require.Error(err)
	require.Contains(errors.FlattenDetails(err), "Position")

	// Test a clean reset
	require.NoError(sq.Reset(ctx, &ResetOptions{
		Container: ctr,
	}))
}
