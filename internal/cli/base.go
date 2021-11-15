package cli

import (
	"context"

	"github.com/hashicorp/go-hclog"
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
