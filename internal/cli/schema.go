package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/sqlbuild"
)

type SchemaCommand struct {
	*baseCommand

	sqlDir string
}

func (c *SchemaCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Determine our directory
	sqlDir, err := filepath.Abs(c.sqlDir)
	if err != nil {
		return c.exitError(fmt.Errorf("Error expanding sql directory: %w", err))
	}
	rootDir, rootFile := filepath.Split(sqlDir)
	if len(rootDir) > 0 && rootDir[len(rootDir)-1] == filepath.Separator {
		rootDir = rootDir[:len(rootDir)-1]
	}

	// Build to our output
	if err := sqlbuild.Build(&sqlbuild.Config{
		Output: os.Stdout,
		FS:     os.DirFS(rootDir),
		Root:   rootFile,
		Logger: c.Log.Named("sqlbuild"),
	}); err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *SchemaCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		f := sets.NewSet("Command Options")

		f.StringVar(&flag.StringVar{
			Name:    "sqldir",
			Target:  &c.sqlDir,
			Default: "sql",
			Usage:   "Root directory for SQL files",
		})
	})
}

func (c *SchemaCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *SchemaCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *SchemaCommand) Synopsis() string {
	return "Build SQL schema"
}

func (c *SchemaCommand) Help() string {
	return formatHelp(`
Usage: squire schema [options]

  Output the full SQL schema.

  This builds the SQL schema from the .sql files in your "sql/" directory
  and outputs it to stdout. This is NOT reading the currently deployed
  schema from any database.

` + c.Flags().Help())
}
