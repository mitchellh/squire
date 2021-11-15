package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/squire/internal/pkg/flag"
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
