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

	// This should error because our SQL directory doesn't exist.
	require.Equal(code, 1)
	require.Empty(out.String())
	require.NotEmpty(err.String())
}

/*
TODO: gotta create flag to set sql dir
func TestSchema_good(t *testing.T) {
	require := require.New(t)

	// Get our working directory before
	wd, err := os.Getwd()
	require.NoError(err)

	base, outBuf, errBuf, finalize := testCLI(t)
	defer finalize()

	// Build our command
	cmd := &SchemaCommand{
		baseCommand: base,
	}

	// No arguments
	code := cmd.Run([]string{
		"-sqldir", filepath.Join(wd, "testdata/schema-good"),
		"-w=false",
	})
	finalize()

	// This should error because our SQL directory doesn't exist.
	require.Equal(0, code)
	require.Empty(errBuf.String())
	require.NotEmpty(outBuf.String())
}
*/
