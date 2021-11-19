package config

import (
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
