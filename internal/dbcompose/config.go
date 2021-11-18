package dbcompose

import (
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
)

// Config is the primary configuration structure. This should not be constructed
// directly. Instead, you should use the constructor methods.
type Config struct {
	logger  hclog.Logger
	project *types.Project

	// populated by init
	service *types.ServiceConfig
	connURI string
}

// New initializes a new container configuration. This will load and validate
// all the settings associated with the specification.
func New(opts ...Option) (*Config, error) {
	// Build our options
	var optsVal options
	for _, o := range opts {
		o(&optsVal)
	}
	optsVal.SetDefaults()

	// Build our config
	cfg := &Config{logger: optsVal.Logger}
	L := cfg.logger

	// Try to find a docker compose file
	if len(optsVal.Paths) > 0 {
		L.Info("looking for compose file", "paths", optsVal.Paths)
		for _, p := range optsVal.Paths {
			LL := L.With("path", p)
			_, err := os.Stat(p)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				LL.Error("error reading compose file", "err", err)
				return nil, err
			}

			// This compose file should work, so we read it.
			cfg.project, err = loadFromFile(p)
			if err != nil {
				return nil, err
			}
		}

		// If our project is still nil, and we have no default, then we error.
		if cfg.project == nil && optsVal.Default == nil {
			return nil, errors.WithDetailf(
				errors.Newf("failed to find a Docker Compose file: %v", optsVal.Paths),
				strings.TrimSpace(errDetailNoFile),
				optsVal.Paths,
			)
		}
	}

	// If our project is nil we set a default
	if cfg.project == nil {
		cfg.project = optsVal.Default
	}

	// If our project is STILL nil then its an error. It shouldn't be
	// possible to reach this point with valid usage.
	if cfg.project == nil {
		panic("nil project, no default?")
	}

	// Initialize, validate
	if err := cfg.init(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Project returns the project for this config.
func (c *Config) Project() *types.Project {
	return c.project
}

// Service returns the service name that is being used.
func (c *Config) Service() string {
	return c.service.Name
}

// ConnURI is a URI for connecting to PostgreSQL. This should be able to
// be used with most PostgreSQL clients.
func (c *Config) ConnURI() string {
	return c.connURI
}
