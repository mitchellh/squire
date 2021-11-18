package dbcontainer

import (
	"context"

	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/squire/internal/dbcompose"
)

type Container struct {
	logger  hclog.Logger
	compose composeapi.Service
	config  *dbcompose.Config
}

// New creates a new Container instance to represent a new or existing
// desired container. For new containers, this will not physically start
// the container until Up is called.
func New(opts ...Option) (*Container, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	// Initialize our API service so we can run compose lifecycle ops
	dockerCli, err := dockerCli()
	if err != nil {
		return nil, err
	}
	api := compose.NewComposeService(dockerCli.Client(), dockerCli.ConfigFile())

	return &Container{
		logger:  cfg.Logger,
		compose: api,
		config:  cfg.ComposeConfig,
	}, nil
}

// ConnURI returns the connection string in URI format. This can be used
// with most PostgreSQL API clients. This can't fail because validation of
// the connection information is precomputed in New and any errors are
// reported then. Note its still possible for the connection itself to fail
// if the container isn't running, invalid information was provided, etc.
func (c *Container) ConnURI() string {
	return c.config.ConnURI()
}

// Up starts the container. If it is already running, this does nothing.
func (c *Container) Up(ctx context.Context) error {
	c.logger.Info("up", "service", c.config.Service())
	return c.compose.Up(ctx, c.config.Project(), composeapi.UpOptions{
		Create: composeapi.CreateOptions{
			Services: []string{c.config.Service()},
		},

		Start: composeapi.StartOptions{},
	})
}

// Down stops the container and removes any data associated with it.
// Note that this will destroy ALL services in the docker-compose project,
// we have no way to filter that out.
func (c *Container) Down(ctx context.Context) error {
	p := c.config.Project()
	c.logger.Info("down", "project", p.Name)
	return c.compose.Down(ctx, p.Name, composeapi.DownOptions{
		Project: p,
		Volumes: true,
	})
}
