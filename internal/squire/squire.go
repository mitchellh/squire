package squire

import (
	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/squire/internal/config"
)

// Squire is the primary struct to perform Squire operations.
type Squire struct {
	logger hclog.Logger
	config *config.Config
}

// Option is used to create a new Squire.
type Option func(*optstruct)

// New creates a new Squire instance.
func New(opts ...Option) (*Squire, error) {
	// Initial config
	var optval optstruct
	for _, opt := range opts {
		opt(&optval)
	}

	// Defaults
	if optval.Logger == nil {
		optval.Logger = hclog.L().Named("squire")
	}
	if optval.Config == nil {
		cfg, err := config.New()
		if err != nil {
			return nil, err
		}

		optval.Config = cfg
	}

	return &Squire{
		logger: optval.Logger,
		config: optval.Config,
	}, nil
}

// WithConfig specifies the configuration to use.
func WithConfig(v *config.Config) Option {
	return func(s *optstruct) {
		s.Config = v
	}
}

// WithLogger specifies the logger to use.
func WithLogger(v hclog.Logger) Option {
	return func(s *optstruct) {
		s.Logger = v
	}
}

type optstruct struct {
	Config *config.Config
	Logger hclog.Logger
}
