package dbcontainer

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/hashicorp/go-hclog"
	_ "github.com/jackc/pgx/v4/stdlib"

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

// Clone clones the container settings. This does not copy any data.
// This does not create or start the cloned container.
func (c *Container) Clone(n string) (*Container, error) {
	cfg2, err := c.config.Clone(n)
	if err != nil {
		return nil, err
	}

	return &Container{
		logger:  c.logger,
		compose: c.compose,
		config:  cfg2,
	}, nil
}

// Config returns the underlying compose configuration.
func (c *Container) Config() *dbcompose.Config {
	return c.config
}

// ConnURI returns the connection string in URI format. This can be used
// with most PostgreSQL API clients. This can't fail because validation of
// the connection information is precomputed in New and any errors are
// reported then. Note its still possible for the connection itself to fail
// if the container isn't running, invalid information was provided, etc.
func (c *Container) ConnURI() string {
	return c.config.ConnURI()
}

// Conn establishes a connection to the database and waits for the DB
// to become ready before returning.
//
// This creates a NEW connection. Callers must close the connection when
// they're done. This only works if the container is running.
func (c *Container) Conn(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("pgx", c.ConnURI())
	if err != nil {
		return nil, err
	}

	// Wrap our context in its own timeout that is reasonable. For
	// container-based bootups, we expect to come up within a minute.
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// Wait for the db to come to life.
	err = backoff.Retry(func() error {
		return db.Ping()
	}, backoff.WithContext(
		backoff.NewConstantBackOff(15*time.Millisecond),
		ctx,
	))
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
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

// Status returns the status of the container.
//
// If the container is not created, a non-nil status will be returned
// with State == NotCreated. Other fields in the status in this case
// will be undefined (may be empty or not, but are meaningless and
// should not be depended on).
func (c *Container) Status(ctx context.Context) (*Status, error) {
	p := c.config.Project()
	containers, err := c.compose.Ps(ctx, p.Name, composeapi.PsOptions{
		Services: []string{c.config.Service()},
	})
	if err != nil {
		return nil, err
	}

	// Not created
	if len(containers) == 0 {
		return &Status{State: NotCreated}, nil
	}

	// No idea what to make of this.
	if len(containers) > 1 {
		c.logger.Warn("more than one container on status", "containers", containers)
	}

	c0 := containers[0]
	return &Status{
		ID:    c0.ID,
		Name:  c0.Name,
		State: State(strings.ToLower(c0.State)),
	}, nil
}
