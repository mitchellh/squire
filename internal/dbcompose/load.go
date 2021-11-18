package dbcompose

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
	composecli "github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
)

// loadFromFile loads a project from a file.
func loadFromFile(path string) (*types.Project, error) {
	opts, err := composecli.NewProjectOptions(
		[]string{path},
	)
	if err != nil {
		return nil, err
	}

	return composecli.ProjectFromOptions(opts)
}

// init should be called once project and logger are set to populate
// the remaining metadata. This also validates that the configuration is
// valid.
func (c *Config) init() error {
	// Grab our service
	svc, err := service(c.project)
	if err != nil {
		return err
	}
	c.service = svc

	// Get connection URL
	uri, err := connURI(svc)
	if err != nil {
		return err
	}
	c.connURI = uri

	return nil
}

// service returns the service configuration for the service representing
// the database.
//
// Precondition: c.project != nil
func service(p *types.Project) (*types.ServiceConfig, error) {
	var result *types.ServiceConfig
	for _, s := range p.Services {
		if len(s.Extensions) == 0 {
			continue
		}

		_, ok := s.Extensions[extName]
		if !ok {
			continue
		}

		// Multiple squire services is an error, for now.
		if result != nil {
			return nil, errors.WithDetailf(
				errors.New("multiple services marked with x-squire"),
				strings.TrimSpace(errDetailMultiService),
				result.Name, s.Name,
			)
		}

		// We have to copy so we don't reuse the s reference
		v := s
		result = &v
	}

	// If we found, return
	if result != nil {
		return result, nil
	}

	return nil, errors.WithDetailf(
		errors.New("failed to find PostgreSQL service in compose specification"),
		strings.TrimSpace(errDetailNoService),
	)
}

// connURI determines the connection URI.
func connURI(svc *types.ServiceConfig) (string, error) {
	// Determine our port
	pgPort, err := pgPort(svc)
	if err != nil {
		return "", err
	}

	// Determine our database name
	pgDB, err := pgDB(svc)
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

// pgPort determines the port for the database on the host.
//
// This works by looking for a port forward from port 5432 (the default pg port)
// to anything on the host.
func pgPort(svc *types.ServiceConfig) (uint32, error) {
	targetPort := uint32(pgDefaultPort)

	// First load from our extension
	ext, err := parseExtension(svc)
	if err != nil {
		return 0, err
	}
	if ext.TargetPort > 0 {
		targetPort = ext.TargetPort
	}

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
func pgDB(svc *types.ServiceConfig) (string, error) {
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

const (
	// pgDefaultPort is the port for the PostgreSQL database.
	pgDefaultPort = 5432

	// pgDBEnv is the env var in the container that specifies the DB name.
	pgDBEnv = "POSTGRES_DB"
)
