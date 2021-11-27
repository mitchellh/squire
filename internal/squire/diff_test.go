package squire

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mitchellh/squire/internal/config"
)

func TestDiff(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	// Build our config
	cfg, err := config.New(config.FromString(
		`sql_dir: "testdata/diff-1"`))
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

	// Reset it
	require.NoError(sq.Reset(ctx, &ResetOptions{
		Container: ctr,
	}))

	// Create our new changed config
	cfg2, err := config.New(config.FromString(
		`sql_dir: "testdata/diff-2"`))
	require.NoError(err)
	sq2, err := New(WithConfig(cfg2))
	require.NoError(err)

	// Diff!
	var out bytes.Buffer
	require.NoError(sq2.Diff(ctx, &DiffOptions{Output: &out}))
	require.NotEmpty(out.String())
	t.Logf("output: %s", out.String())
	out1 := out.String()

	// Diff again with verification
	out.Reset()
	require.NoError(sq2.Diff(ctx, &DiffOptions{Output: &out, Verify: true}))
	require.NotEmpty(out.String())
	require.Equal(out1, out.String())
}
