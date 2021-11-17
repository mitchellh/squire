package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/copy"
	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/sqlbuild"
)

type SchemaCommand struct {
	*baseCommand

	sqlDir string
	write  bool
}

func (c *SchemaCommand) Run(args []string) int {
	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Determine our directory. We want an absolute directory so we can
	// put together the fs.FS implementation.
	sqlDir, err := filepath.Abs(c.sqlDir)
	if err != nil {
		return c.exitError(fmt.Errorf("Error expanding sql directory: %w", err))
	}
	rootDir, rootFile := filepath.Split(sqlDir)
	if len(rootDir) > 0 && rootDir[len(rootDir)-1] == filepath.Separator {
		// Strip the trailining filepath separator.
		rootDir = rootDir[:len(rootDir)-1]
	}

	// Our build output is stdout.
	var buildOutput io.Writer = os.Stdout

	// If we're writing, we want to tee in a temporary file that we'll
	// replace our schema with on success. We don't write directly to
	// the schema because we don't want to corrupt it.
	var schemaFile *os.File
	if c.write {
		schemaFile, err = ioutil.TempFile("", "squire")
		if err != nil {
			return c.exitError(fmt.Errorf(
				"Error creating temporary file for schema: %w", err))
		}
		defer os.Remove(schemaFile.Name())
		buildOutput = io.MultiWriter(buildOutput, schemaFile)
	}

	// Build to our output
	if err := sqlbuild.Build(&sqlbuild.Config{
		Output: buildOutput,
		FS:     os.DirFS(rootDir),
		Root:   rootFile,
		Logger: c.Log.Named("sqlbuild"),
		Metadata: map[string]string{
			"Generation Time": time.Now().Format(time.UnixDate),
		},
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
		final := filepath.Join(sqlDir, "schema.sql")

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

		f.StringVar(&flag.StringVar{
			Name:    "sqldir",
			Target:  &c.sqlDir,
			Default: "sql",
			Usage:   "Root directory for SQL files",
		})

		f.BoolVar(&flag.BoolVar{
			Name:    "write",
			Target:  &c.write,
			Default: true,
			Usage:   "Write the SQL schemas to <sqldir>/schema.sql",
			Aliases: []string{"w"},
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
