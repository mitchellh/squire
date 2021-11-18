package dbcontainer

import (
	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/squire/internal/dbcompose"
)

// Option configures the container.
type Option func(*config)

// WithLogger specifies a logger to use. If this is not set, we will use
// the default logger with the name "container".
func WithLogger(v hclog.Logger) Option {
	return func(cfg *config) {
		cfg.Logger = v
	}
}

// WithCompose specifies the compose configuration.
func WithCompose(v *dbcompose.Config) Option {
	return func(cfg *config) {
		cfg.ComposeConfig = v
	}
}

// config is the configuration for the container. This must be constructed
// and modified through various Option functions rather than directly.
type config struct {
	Logger        hclog.Logger
	ComposeConfig *dbcompose.Config
}

func newConfig(opts ...Option) (*config, error) {
	// Load our options
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	// Default logger
	if cfg.Logger == nil {
		cfg.Logger = hclog.L().Named("container")
	}

	return &cfg, nil
}
