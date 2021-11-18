package dbcompose

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
)

// Option is used to configure the New function.
type Option func(*options)

// WithPath adds a path to the search path for a docker compose file. If
// it isn't found, it is not an error if Default is set. If Default is NOT
// set, then it will error.
func WithPath(v string) Option {
	return func(final *options) {
		final.Paths = append(final.Paths, v)
	}
}

// WithLogger specifies a logger to use. If this is not set, we will use
// the default logger with the name "container".
func WithLogger(v hclog.Logger) Option {
	return func(final *options) {
		final.Logger = v
	}
}

// WithDefault specifies a default project to use if no path resolves.
func WithDefault(v *types.Project) Option {
	return func(final *options) {
		final.Default = v
	}
}

type options struct {
	Logger  hclog.Logger
	Paths   []string
	Default *types.Project
}

// SetDefaults should be called to set all default values on options.
func (v *options) SetDefaults() {
	if v.Logger == nil {
		v.Logger = hclog.L().Named("dbcompose")
	}
}
