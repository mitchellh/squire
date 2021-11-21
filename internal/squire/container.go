package squire

import (
	"github.com/mitchellh/squire/internal/config"
	"github.com/mitchellh/squire/internal/dbcompose"
	"github.com/mitchellh/squire/internal/dbcontainer"
	"github.com/mitchellh/squire/internal/dbdefault"
)

// Container returns the primary dev container for this instance. The
// container instance can then be further used to get access to clones.
func (s *Squire) Container() (*dbcontainer.Container, error) {
	// Look for a docker compose yaml file in parent directories
	composePath, err := config.FindPath("", "docker-compose.yml")
	if err != nil {
		return nil, err
	}
	if composePath == "" {
		composePath = "docker-compose.yml"
	}

	// Build our config
	cfg, err := dbcompose.New(
		dbcompose.WithLogger(s.logger.Named("compose")),
		dbcompose.WithDefault(dbdefault.Project()),
		dbcompose.WithPath(composePath),
	)
	if err != nil {
		return nil, err
	}

	// Build our container
	return dbcontainer.New(
		dbcontainer.WithLogger(s.logger.Named("container")),
		dbcontainer.WithCompose(cfg),
	)
}
