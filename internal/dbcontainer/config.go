package dbcontainer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
	composecli "github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	dockercommand "github.com/docker/cli/cli/command"
	dockercliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/hashicorp/go-hclog"
)

// Option configures the container.
type Option func(*config) error

// WithComposeFile reads the docker-compose YAML file to populate the project.
func WithComposeFile(path string) Option {
	return func(dst *config) error {
		opts, err := composecli.NewProjectOptions(
			[]string{path},
		)
		if err != nil {
			return err
		}

		proj, err := composecli.ProjectFromOptions(opts)
		if err != nil {
			return err
		}

		dst.Project = proj
		return nil
	}
}

// WithService is the service name in the compose specification to use
// as the template for launching the database.
func WithService(n string) Option {
	return func(dst *config) error {
		dst.Service = n
		return nil
	}
}

// WithLogger specifies a logger to use. If this is not set, we will use
// the default logger with the name "container".
func WithLogger(v hclog.Logger) Option {
	return func(dst *config) error {
		dst.Logger = v
		return nil
	}
}

// config is the configuration for the container. This must be constructed
// and modified through various Option functions rather than directly.
type config struct {
	Logger  hclog.Logger
	Project *types.Project
	Service string
}

func newConfig(opts ...Option) (*config, error) {
	// Load our options
	var cfg config
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	// Default logger
	if cfg.Logger == nil {
		cfg.Logger = hclog.L().Named("container")
	}

	return &cfg, nil
}

// apiService returns the compose api.Service implementation so that
// lifecycle operations such as Up, Down, etc. can start being called
// on the project.
func (c *config) apiService() (api.Service, error) {
	// Initialize the docker CLI 💀. The compose library needs this to operate.
	cli, err := dockercommand.NewDockerCli()
	if err != nil {
		return nil, err
	}
	if err := cli.Initialize(dockercliflags.NewClientOptions()); err != nil {
		return nil, err
	}

	return compose.NewComposeService(
		cli.Client(),
		cli.ConfigFile(),
	), nil
}

// service returns the service configuration for the service representing
// the database.
//
// Precondition: c.Project != nil
func (c *config) service() (*types.ServiceConfig, error) {
	for _, s := range c.Project.Services {
		if strings.EqualFold(s.Name, c.Service) {
			return &s, nil
		}
	}

	return nil, errors.WithDetailf(
		errors.Newf("failed to find database service %q in compose specification", c.Service),
		strings.TrimSpace(errDetailNoService),
		c.Service,
	)
}

// pgPort determines the port for the database on the host.
//
// This works by looking for a port forward from port 5432 (the default pg port)
// to anything on the host.
func (c *config) pgPort(svc *types.ServiceConfig) (uint32, error) {
	// Pulling this out in case we parameterize it later.
	const targetPort = pgDefaultPort

	for _, p := range svc.Ports {
		if p.Target == targetPort {
			return p.Published, nil
		}
	}

	return 0, errors.WithDetailf(
		errors.Newf("failed to determine PostgreSQL port for service %q", svc.Name),
		strings.TrimSpace(errDetailNoPort),
		targetPort,
	)
}

// pgDB determines the database name.
func (c *config) pgDB(svc *types.ServiceConfig) (string, error) {
	// First load from our extension
	ext, err := parseExtension(svc)
	if err != nil {
		return "", err
	}
	if ext.DB != "" {
		return ext.DB, nil
	}

	// Get the DB from env vars
	if len(svc.Environment) > 0 {
		v, ok := svc.Environment[pgDBEnv]
		if ok && v != nil {
			return *v, nil
		}
	}

	return "", errors.WithDetailf(
		errors.Newf("failed to determine PostgreSQL database name for service %q", svc.Name),
		strings.TrimSpace(errDetailNoDB),
		svc.Name,
	)
}

// connURI determines the connection URI.
func (c *config) connURI(svc *types.ServiceConfig) (string, error) {
	// Determine our port
	pgPort, err := c.pgPort(svc)
	if err != nil {
		return "", err
	}

	// Determine our database name
	pgDB, err := c.pgDB(svc)
	if err != nil {
		return "", err
	}

	var u url.URL
	u.Scheme = "postgres"
	u.Host = fmt.Sprintf("localhost:%d", pgPort)
	u.User = url.User("postgres")
	u.Path = pgDB

	return u.String(), nil
}

const (
	// pgDefaultPort is the port for the PostgreSQL database.
	pgDefaultPort = 5432

	// pgDBEnv is the env var in the container that specifies the DB name.
	pgDBEnv = "POSTGRES_DB"
)

const (
	errDetailNoService = `
Squire requires a database service named %q is present in the Docker Compose
file. This is used as the template to launch your database for dev and test.
Please fix this by introducing the database service:

services:
  %[1]s:
    ...
`

	errDetailNoDB = `
Squire needs to know the name of the database within PostgreSQL to use.
TODO
`

	errDetailDBString = `
The database name specified by the "x-squire.db" field must be a string.
For example:

services:
  %[1]s:
    x-squire:
      db: "my-db"
`

	errDetailNoPort = `
Squire needs to know the port to use to communicate with the PostgreSQL
container on the host, but failed to determine that port. The port is detected
by looking for a port forwarding to the target port %d. You can fix this
by introducing a forwarded port in your compose specification:

services:
  <your db service>:
    ports:
      - "<any host port>:5432"
`
)
