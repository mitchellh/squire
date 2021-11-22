// Package config can load configuration for Squire.
package config

import (
	"bytes"
	stderr "errors"
	"os"
	"os/exec"
	"strings"

	"cuelang.org/go/cue"
	"github.com/cockroachdb/errors"
)

// Config is the final configuration structure for Squire.
type Config struct {
	// Root is the root value for the configuration. This might be nil
	// if the configuration was hand-created.
	Root cue.Value `json:"-"`

	SQLDir string `json:"sql_dir"`

	Dev struct {
		DefaultImage string `json:"default_image"`
	}

	Production struct {
		Mode    string
		Env     string
		Command []string
	}
}

// ProdURL returns the URL to the production database. This will never
// return an empty string with a nil error. This will return an error if the
// production URL could not be determined. An empty string error will be
// ErrProdNotFound
func (c *Config) ProdURL() (string, error) {
	switch c.Production.Mode {
	case "env":
		return c.prodEnv()

	case "exec":
		return c.prodExec()

	default:
		return "", errors.WithDetail(
			errors.Newf("invalid production mode: %q", c.Production.Mode),
			strings.TrimSpace(errDetailMode),
		)
	}
}

func (c *Config) prodEnv() (string, error) {
	v := os.Getenv(c.Production.Env)
	if v == "" {
		return "", ErrProdNotFound
	}

	return v, nil
}

func (c *Config) prodExec() (string, error) {
	args := c.Production.Command
	if len(args) == 0 {
		return "", errors.New("'production.command' must be non-empty")
	}

	var bufout, buferr bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &bufout
	cmd.Stderr = &buferr
	if err := cmd.Run(); err != nil {
		return "", errors.WithDetailf(
			errors.Newf("error executing command for production URL: %v", args),
			strings.TrimSpace(errDetailExec),
			bufout.String(), buferr.String(),
		)
	}

	bs := bufout.Bytes()
	if idx := bytes.IndexByte(bs, '\n'); idx >= 0 {
		bs = bs[:idx]
	}

	return string(bs), nil
}

var ErrProdNotFound = stderr.New("production connection URL is empty")

const (
	errDetailMode = `
The only valid modes to acquire a production connection URL are "env" and "exec".
Run "squire config -default -full" to see the full default configuration
including documentation.
`

	errDetailExec = `
There was an error while executing the configured command to load the
production database URL. The stdout and stderr is below:

stdout:

%[1]s

stderr:

%[2]s
`
)
