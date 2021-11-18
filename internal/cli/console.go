package cli

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/exec"
	"github.com/mitchellh/squire/internal/pkg/flag"
)

type ConsoleCommand struct {
	*baseCommand

	sqlDir string
	write  bool
}

func (c *ConsoleCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Get our container
	ctr, err := c.container()
	if err != nil {
		return c.exitError(err)
	}

	// Get the URI
	uri := ctr.ConnURI()

	// Look for psql
	argv0, err := osexec.LookPath("psql")
	if err != nil {
		return c.exitError(errors.WithDetail(
			errors.New("psql not found"),
			strings.TrimSpace(errDetailNoPsql),
		))
	}

	// Launch it
	fmt.Printf("==> Connecting to: %s\n", uri)
	if err := exec.Exec(argv0, []string{argv0, uri}, os.Environ()); err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *ConsoleCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *ConsoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ConsoleCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *ConsoleCommand) Synopsis() string {
	return "Enter a psql console for the dev database"
}

func (c *ConsoleCommand) Help() string {
	return formatHelp(`
Usage: squire console [options]

  Enter a "psql" interactive console for the dev database.

  This requires that the database is running (from calling "up") and
  that your system has "psql" available.

` + c.Flags().Help())
}

const (
	errDetailNoPsql = `
"psql" was not found installed on your system. "squire console" requires
"psql" to be available. This is usually found by installing the
default PostgreSQL package for your operating system.
`
)
