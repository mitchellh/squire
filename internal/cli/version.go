package cli

import (
	"fmt"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/version"
)

type VersionCommand struct {
	*baseCommand
}

func (c *VersionCommand) Run(args []string) int {
	out := version.Info().String()
	fmt.Printf("%s %s", cliName, out)

	return 0
}

func (c *VersionCommand) Flags() *flag.Sets {
	return nil
}

func (c *VersionCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *VersionCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the version of this Waypoint CLI"
}

func (c *VersionCommand) Help() string {
	return formatHelp(`
Usage: waypoint version
  Prints the version information for Waypoint.
  This command will show the version of the current Waypoint CLI. If
  the CLI is configured to communicate to a Waypoint server, the server
  version will also be shown.
`)
}
