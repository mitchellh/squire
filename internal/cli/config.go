package cli

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
)

type ConfigCommand struct {
	*baseCommand

	json bool
}

func (c *ConfigCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	if c.json {
		bs, err := c.Config.Root.MarshalJSON()
		if err != nil {
			return c.exitError(err)
		}

		fmt.Println(string(bs))
		return 0
	}

	// Get our config
	node := c.Config.Root.Syntax(
		cue.Concrete(true),
		cue.Docs(true),
	)
	bs, err := format.Node(node)
	if err != nil {
		return c.exitError(err)
	}

	fmt.Println(string(bs))
	return 0
}

func (c *ConfigCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		f := sets.NewSet("Command Options")

		f.BoolVar(&flag.BoolVar{
			Name:    "json",
			Target:  &c.json,
			Default: false,
			Usage:   "Write the configuration in JSON format.",
		})
	})
}

func (c *ConfigCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ConfigCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *ConfigCommand) Synopsis() string {
	return "Output the current configuration"
}

func (c *ConfigCommand) Help() string {
	return formatHelp(`
Usage: squire config [options]

  Output the current Squire configuration.

  This will output the current squire configuration, or the default
  configuration if not explicit configuration file is detected. This can
  be used to inspect the defaults or verify that your configuration changes
  are applying.

  The default format will be Cue with comments (documentation) included.
  This is a fully valid configuration file you can start with. The "--json"
  flag can be specified to output in JSON.

  The configuration file should be saved to .squire, .squire.cue, or
  .squire.json (with only the final filename being json-formatted and
  the others being in Cue). Only one file will be loaded, in the order
  given.

` + c.Flags().Help())
}
