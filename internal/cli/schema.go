package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/copy"
	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type SchemaCommand struct {
	*baseCommand

	write     bool
	tests     bool
	testsOnly bool
}

func (c *SchemaCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// If we're rendering tests, do not write.
	if c.tests {
		c.write = false
	}

	// Our build output is stdout.
	var buildOutput io.Writer = os.Stdout

	// If we're writing, we want to tee in a temporary file that we'll
	// replace our schema with on success. We don't write directly to
	// the schema because we don't want to corrupt it.
	var schemaFile *os.File
	if c.write {
		var err error
		schemaFile, err = ioutil.TempFile("", "squire")
		if err != nil {
			return c.exitError(fmt.Errorf(
				"Error creating temporary file for schema: %w", err))
		}
		defer os.Remove(schemaFile.Name())
		buildOutput = io.MultiWriter(buildOutput, schemaFile)
	}

	// Build to our output
	if err := c.Squire.Schema(&squire.SchemaOptions{
		Output:    buildOutput,
		Tests:     c.tests,
		TestsOnly: c.testsOnly,
	}); err != nil {
		return c.exitError(err)
	}

	// If we are writing the schema, copy it over now
	if schemaFile != nil {
		// Close to flush all our data
		if err := schemaFile.Close(); err != nil {
			return c.exitError(fmt.Errorf(
				"Error closing temporary file for schema: %w", err))
		}

		// Our final path is the sqldir
		final := filepath.Join(c.Config.SQLDir, "schema.sql")

		// Copy our file
		if err := copy.CopyFile(schemaFile.Name(), final); err != nil {
			return c.exitError(fmt.Errorf(
				"Error writing schema: %w", err))
		}
	}

	return 0
}

func (c *SchemaCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		f := sets.NewSet("Command Options")

		f.BoolVar(&flag.BoolVar{
			Name:    "write",
			Target:  &c.write,
			Default: true,
			Usage:   "Write the SQL schemas to <sqldir>/schema.sql",
			Aliases: []string{"w"},
		})

		f.BoolVar(&flag.BoolVar{
			Name:    "test",
			Target:  &c.tests,
			Default: false,
			Usage: "Include the test SQL files (ending in _test.sql). " +
				"This implies -write=false.",
		})

		f.BoolVar(&flag.BoolVar{
			Name:    "test-only",
			Target:  &c.testsOnly,
			Default: false,
			Usage: "Only show test SQL files. This requires -test. " +
				"This implies -write=false.",
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

  By default, this only includes non-test SQL files (SQL files that do
  not end in "_test.sql"). The "-test" flag can be specified to include
  test SQL files and the "-test-only" flag can be specified to ONLY
  show test SQL files.

` + c.Flags().Help())
}
