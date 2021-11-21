package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/squire/internal/config"
	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

// baseCommand is embedded in all commands to provide common logic and data.
//
// The unexported values are not available until after Init is called. Some
// values are only available in certain circumstances, read the documentation
// for the field to determine if that is the case.
type baseCommand struct {
	// Ctx is the base context for the command. It is up to commands to
	// utilize this context so that cancellation works in a timely manner.
	Ctx context.Context

	// Log is the logger to use.
	Log hclog.Logger

	//---------------------------------------------------------------
	// Set after Init

	Config *config.Config
	Squire *squire.Squire
}

// Close implements io.Closer. This should be called to gracefully clean up
// any resources created by the CLI command.
func (c *baseCommand) Close() error {
	// Nothing today, but we expect this to be called so we can add stuff later.
	return nil
}

// Init initializes the command. This should be called early no matter what.
// You can control what is done by using the options.
func (c *baseCommand) Init(opts ...Option) error {
	baseCfg := baseConfig{}
	for _, opt := range opts {
		opt(&baseCfg)
	}

	// Parse flags
	args := baseCfg.Args
	if baseCfg.Flags != nil {
		if err := baseCfg.Flags.Parse(baseCfg.Args); err != nil {
			return err
		}

		args = baseCfg.Flags.Args()
		if v := baseCfg.FlagOutArgs; v != nil {
			*v = args
		}
	}

	// Check for flags after args
	if err := checkFlagsAfterArgs(args, baseCfg.Flags); err != nil {
		return err
	}

	// Load our config
	if err := c.loadConfig(); err != nil {
		return err
	}

	// Initialize squire
	sq, err := squire.New(
		squire.WithConfig(c.Config),
		squire.WithLogger(c.Log.Named("squire")),
	)
	if err != nil {
		return err
	}
	c.Squire = sq

	return nil
}

// loadConfig loads the configuration and sets it on the base.
func (c *baseCommand) loadConfig() error {
	var opts []config.Option

	// Look for a config file
	path, err := config.FindPath("", config.Filename)
	if err != nil {
		return err
	}
	if path == "" {
		path, err = config.FindPath("", config.Filename+".cue")
		if err != nil {
			return err
		}
	}
	if path == "" {
		path, err = config.FindPath("", config.Filename+".json")
		if err != nil {
			return err
		}
	}

	// If we have a config file load it
	if path != "" {
		opts = append(opts, config.FromFile(path))
	}

	// Load
	c.Config, err = config.New(opts...)
	if err != nil {
		return err
	}

	return nil
}

// exitError should be called by commands to exit with an error.
func (c *baseCommand) exitError(err error) int {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	return 1
}

// flagSet creates the flags for this command. The callback should be used
// to configure the set with your own custom options.
func (c *baseCommand) flagSet(bit flagSetBit, f func(*flag.Sets)) *flag.Sets {
	set := flag.NewSets()

	if f != nil {
		f(set)
	}

	return set
}

// flagSetBit is used with baseCommand.flagSet
type flagSetBit uint

const (
	flagSetDefault flagSetBit = 1 << iota
)

// Option is used to configure Init on baseCommand.
type Option func(c *baseConfig)

// WithArgs sets the arguments to the command that are used for parsing.
// Remaining arguments can be accessed using your flag set and asking for Args.
// Example: c.Flags().Args().
func WithArgs(args []string) Option {
	return func(c *baseConfig) { c.Args = args }
}

// WithFlags sets the flags that are supported by this command.
//
// outArgs can be used to specify where remaining positional arguments
// are written to. This can be nil and no positional arguments will be
// recorded.
func WithFlags(f *flag.Sets, outArgs *[]string) Option {
	return func(c *baseConfig) {
		c.Flags = f
		c.FlagOutArgs = outArgs
	}
}

type baseConfig struct {
	Args        []string
	Flags       *flag.Sets
	FlagOutArgs *[]string
}
