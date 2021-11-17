package sqlbuild

import (
	"bytes"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	// cases to run, we manually go through them because
	// in the future we'll have cases that set specific
	// options and so on.
	cases := []string{
		"build",
	}

	for _, n := range cases {
		t.Run(n, func(t *testing.T) {
			g := goldie.New(t,
				goldie.WithNameSuffix(".golden.sql"),
			)

			var buf bytes.Buffer
			require.NoError(t, Build(&Config{
				Output: &buf,
				FS:     os.DirFS("testdata"),
				Root:   n,
				Logger: hclog.New(&hclog.LoggerOptions{
					Level: hclog.Debug,
				}),
			}))

			g.Assert(t, n, buf.Bytes())
		})
	}
}

func TestBuild_noExist(t *testing.T) {
	var buf bytes.Buffer
	require.Error(t, Build(&Config{
		Output: &buf,
		FS:     os.DirFS("testdata"),
		Root:   "thisdoesntexist",
		Logger: hclog.New(&hclog.LoggerOptions{
			Level: hclog.Debug,
		}),
	}))

	// On error, we should not output anything
	require.Empty(t, buf.String())
}
