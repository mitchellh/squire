package dbcontainer

import (
	"testing"

	"github.com/stretchr/testify/require"
	//"github.com/davecgh/go-spew/spew"
)

func TestFromComposeFile(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := newConfig(
		WithComposeFile("testdata/compose-v2.yml"),
		WithService("postgres"),
	)
	require.NoError(err)
	require.Len(cfg.Project.Services, 1)

	// Verify our service is loaded
	svc, err := cfg.service()
	require.NoError(err)
	require.NotNil(svc)

	// Verify our port
	port, err := cfg.pgPort(svc)
	require.NoError(err)
	require.Equal(uint32(1234), port)

	// Try api service
	_, err = cfg.apiService()
	require.NoError(err)
}
