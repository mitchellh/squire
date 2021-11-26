package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/version"
)

type VersionCommand struct {
	*baseCommand
}

func (c *VersionCommand) Run(args []string) int {
	out := version.Info().String()
	fmt.Printf("%s %s\n\n", cliName, out)

	// We also perform some basic doctoring here to make notes about our
	// environment. This is helpful to determine how squire will behave.

	fmt.Printf("Squire dependency information:\n")

	// psql
	psql, err := exec.LookPath("psql")
	if err == nil {
		fmt.Printf("✓ psql      (path: %s)\n", psql)
	}
	if errors.Is(err, exec.ErrNotFound) {
		err = nil
		colorError.Println(strings.TrimSpace(errDetailNoPSQL) + "\n")
	}
	if err != nil {
		colorError.Printf("Error looking for psql: %s\n", err)
	}

	// pg_dump
	pgd, err := exec.LookPath("pg_dump")
	if err == nil {
		fmt.Printf("✓ pg_dump   (path: %s)\n", pgd)
	}
	if errors.Is(err, exec.ErrNotFound) {
		err = nil
		colorError.Println(strings.TrimSpace(errDetailNoPGDump) + "\n")
	}
	if err != nil {
		colorError.Printf("Error looking for pg_dump: %s\n", err)
	}

	// pgquarrel
	pgq, err := exec.LookPath("pgquarrel")
	if err == nil {
		fmt.Printf("✓ pgquarrel (path: %s)\n", pgq)
	}
	if errors.Is(err, exec.ErrNotFound) {
		err = nil
		colorError.Println(strings.TrimSpace(errDetailNoPGQuarrel))
	}
	if err != nil {
		colorError.Printf("Error looking for pgquarrel: %s\n", err)
	}

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
	return "Prints the version and environment information"
}

func (c *VersionCommand) Help() string {
	return formatHelp(`
Usage: squire version

  Prints the version information for Squire.

`)
}

const (
	errDetailNoPSQL = `
psql could not be found. The "squire console" command will not work, but
other Squire commands should remain functional.
`

	errDetailNoPGDump = `
pg_dump could not be found. This is used by "squire diff" as an optional
verification mechanism. It is highly recommended you have pg_dump available.
`

	errDetailNoPGQuarrel = `
pgquarrel could not be found. "squire diff" and "squire deploy" will not
work, but other Squire commands should remain functional.
`
)
