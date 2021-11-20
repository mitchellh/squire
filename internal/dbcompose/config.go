package dbcompose

import (
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/copystructure"
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

// Clone creates a "clone" database service. This clone is of the container
// settings and does NOT contain any of the data of the container. The expected
// use case of this is to spin up alternate instances of a database.
//
// The clone has to be given a unique name.
//
// The clone is not persisted to the original compose configuration so if
// the user runs docker-compose down or something outside of Squire then
// it could bring that database down.
func (c *Config) Clone(n string) (*Config, error) {
	// Internal note: I'm not fully sure this is very robust. I did the
	// simplest possible thing to get started and that works for me but I'm
	// fairly certain that if other people ever attempt to use this tool
	// we're going to have to modify how this works. Given I'm unsure how today
	// since this works for me, I'm going to leave this naive implementation
	// to start and we can take a look later.

	// Deep copy our project
	p := copystructure.Must(copystructure.Copy(c.project)).(*types.Project)

	// Modify our project service. What we want to do here is create a new
	// service with a new name that is otherwise identical to our current service.
	svc := copystructure.Must(copystructure.Copy(c.service)).(*types.ServiceConfig)
	svc.Name = fmt.Sprintf("%s-%s", svc.Name, n)

	// Replace the port with some random new port for the db
	if err := pgReplacePort(svc, 0); err != nil {
		return nil, err
	}

	// Replace all ports with only our forwarded port, since we don't
	// want access to anything else.
	portConfig, err := _pgPort(svc)
	if err != nil {
		return nil, err
	}
	svc.Ports = []types.ServicePortConfig{*portConfig}

	// Replace the DB service with our new service. We replace so that
	// there is still exactly one squire service.
	found := false
	for i, s := range p.Services {
		if s.Name == c.service.Name {
			p.Services[i] = *svc
			found = true
			break
		}
	}
	if !found {
		// should never happen
		panic("failed to replace service")
	}

	// Shallow copy ourselves.
	c2 := *c
	c2.project = p

	// Reinitialize
	if err := c2.init(); err != nil {
		return nil, err
	}

	return &c2, nil
}
