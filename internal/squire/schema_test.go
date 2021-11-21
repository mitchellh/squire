package squire

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mitchellh/squire/internal/config"
)

func TestSchema(t *testing.T) {
	require := require.New(t)

	// Build our config
	cfg, err := config.New(config.FromString(
		`sql_dir: "testdata/schema"`))
	require.NoError(err)

	// Build squire
	sq, err := New(WithConfig(cfg))
	require.NoError(err)

	var buf bytes.Buffer
	require.NoError(sq.Schema(&SchemaOptions{
		Output: &buf,
	}))
	require.NotEmpty(buf.String())
}
