package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchema(t *testing.T) {
	require := require.New(t)
	base, out, err, finalize := testCLI(t)
	defer finalize()

	// Build our command
	cmd := &SchemaCommand{
		baseCommand: base,
	}

	// No arguments
	code := cmd.Run([]string{})
	finalize()
	require.Equal(code, 1)
	require.Empty(out.String())
	require.NotEmpty(err.String())
}
