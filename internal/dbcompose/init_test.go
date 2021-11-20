package dbcompose

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPgReplacePort(t *testing.T) {
	require := require.New(t)

	// Load a project
	p, err := loadFromFile("testdata/compose-v2.yml")
	require.NoError(err)

	// Get our usual service
	svc, err := service(p)
	require.NoError(err)

	// Get our typical port
	original, err := pgPort(svc)
	require.NoError(err)
	require.NotEqual(0, original)

	// Replace the port
	require.NoError(pgReplacePort(svc, 7890))

	// Verify
	changed, err := pgPort(svc)
	require.NoError(err)
	require.NotEqual(original, changed)
	require.Equal(uint32(7890), changed)

	// Replace the port with random
	require.NoError(pgReplacePort(svc, 0))
	changed2, err := pgPort(svc)
	require.NoError(err)
	require.NotEqual(original, changed)
	require.NotEqual(changed, changed2)
}
