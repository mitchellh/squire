package dbcompose

import (
	"testing"

	//"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := New(
		WithPath("testdata/compose-v2.yml"),
	)
	require.NoError(err)

	//spew.Dump(cfg.project)

	// Verify our URI
	require.Equal("postgres://postgres@localhost:1234/app-dev", cfg.ConnURI())
}

func Test_noService(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := New(
		WithPath("testdata/compose-no-service.yml"),
	)
	require.Error(err)
	require.Nil(cfg)
	require.Contains(err.Error(), "failed to find")
}

func Test_multiService(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := New(
		WithPath("testdata/compose-multi-service.yml"),
	)
	require.Error(err)
	require.Nil(cfg)
	require.Contains(err.Error(), "multiple")
}

func TestConfigClone(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := New(
		WithPath("testdata/compose-v2.yml"),
	)
	require.NoError(err)

	// Clone it
	cfg2, err := cfg.Clone("test")
	require.NoError(err)

	// Should have different addresses
	require.NotEqual(cfg.ConnURI(), cfg2.ConnURI())
}

/*
func TestFromComposeFile_extension(t *testing.T) {
	require := require.New(t)

	// Load
	cfg, err := newConfig(
		WithComposeFile("testdata/compose-v2-extension-db.yml"),
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

	// Verify our DB
	db, err := cfg.pgDB(svc)
	require.NoError(err)
	require.Equal("foo", db)

	// Try api service
	_, err = cfg.apiService()
	require.NoError(err)
}
*/
