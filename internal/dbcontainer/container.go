package dbcontainer

import (
	"context"

	"github.com/compose-spec/compose-go/types"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/hashicorp/go-hclog"
)

type Container struct {
	logger  hclog.Logger
	compose composeapi.Service
	project *types.Project
	service *types.ServiceConfig
}

// New creates a new Container instance to represent a new or existing
// desired container. For new containers, this will not physically start
// the container until Up is called.
func New(opts ...Option) (*Container, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	// Grab our API service so we can make lifecycle operations happen
	compose, err := cfg.apiService()
	if err != nil {
		return nil, err
	}

	// Grab our primary service
	svc, err := cfg.service()
	if err != nil {
		return nil, err
	}

	// Determine our pg connection information we'd use if the container
	// is running (we don't know or care at this point what the status is).
	// TODO

	return &Container{
		logger:  cfg.Logger,
		compose: compose,
		project: cfg.Project,
		service: svc,
	}, nil
}

// ConnURI returns the connection string in URI format. This can be used
// with most PostgreSQL API clients. This can't fail because validation of
// the connection information is precomputed in New and any errors are
// reported then. Note its still possible for the connection itself to fail
// if the container isn't running, invalid information was provided, etc.
func (c *Container) ConnURI() string {
	return ""
}

// Up starts the container. If it is already running, this does nothing.
func (c *Container) Up(ctx context.Context) error {
	c.logger.Info("up", "service", c.service.Name)
	return c.compose.Up(ctx, c.project, composeapi.UpOptions{
		Create: composeapi.CreateOptions{
			Services: []string{c.service.Name},
		},

		Start: composeapi.StartOptions{},
	})
}

// Down stops the container and removes any data associated with it.
// Note that this will destroy ALL services in the docker-compose project,
// we have no way to filter that out.
func (c *Container) Down(ctx context.Context) error {
	c.logger.Info("down", "project", c.project.Name)
	return c.compose.Down(ctx, c.project.Name, composeapi.DownOptions{
		Project: c.project,
		Volumes: true,
	})
}
