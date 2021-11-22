package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_empty(t *testing.T) {
	require := require.New(t)

	// Empty should load our defaults
	cfg, err := New()
	require.NoError(err)
	require.NotNil(cfg)

	// We should have a default
	require.NotEmpty(cfg.Dev.DefaultImage)
	require.Equal("PGURI", cfg.Production.Env)
}

func TestLoad_file(t *testing.T) {
	require := require.New(t)

	// Empty should load our defaults
	cfg, err := New(FromFile("testdata/default_image.cue"))
	require.NoError(err)
	require.NotNil(cfg)

	// We should have a default
	require.Equal("test", cfg.Dev.DefaultImage)
}

func TestLoad_string(t *testing.T) {
	require := require.New(t)

	// Empty should load our defaults
	cfg, err := New(FromString(`sql_dir: "yo"`))
	require.NoError(err)
	require.NotNil(cfg)

	// We should have a default
	require.Equal("yo", cfg.SQLDir)
}

func TestLoad_prodEnv(t *testing.T) {
	require := require.New(t)

	// Create
	cfg, err := New(
		FromString(`production: mode: "env"`),
		FromString(`production: env: "PGURI"`),
	)
	require.NoError(err)
	require.NotNil(cfg)

	// Set our env
	require.NoError(os.Setenv("PGURI", "foo"))

	url, err := cfg.ProdURL()
	require.NoError(err)
	require.Equal("foo", url)
}
