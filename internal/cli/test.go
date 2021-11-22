package cli

import (
	"database/sql"
	"os"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/mitchellh/go-wordwrap"
	"github.com/posener/complete"

	"github.com/mitchellh/squire/internal/pkg/flag"
	"github.com/mitchellh/squire/internal/squire"
)

type TestCommand struct {
	*baseCommand
}

func (c *TestCommand) Run(args []string) int {
	ctx := c.Ctx

	if err := c.Init(
		WithArgs(args),
		WithFlags(c.Flags(), nil),
	); err != nil {
		return c.exitError(err)
	}

	// Run tests
	if err := c.Squire.TestPGUnit(ctx, &squire.TestPGUnitOptions{
		Callback: c.renderPGUnitResults,
	}); err != nil {
		return c.exitError(err)
	}

	return 0
}

func (c *TestCommand) renderPGUnitResults(rows *sql.Rows) error {
	// Build our table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Get our column names as headers
	var header table.Row
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	for _, col := range cols {
		header = append(header, col)
	}
	t.AppendHeader(header)

	// Get all our row value types ready for scanning
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	var rowVals []reflect.Value
	for _, colType := range colTypes {
		val := reflect.New(colType.ScanType())
		rowVals = append(rowVals, val)
	}

	// Go through each row and scan the result
	scanVal := reflect.ValueOf(rows.Scan)
	for rows.Next() {
		// Scan
		values := scanVal.Call(rowVals)

		// If we have an error, return that. There is only one return val for Scan.
		if raw := values[0].Interface(); raw != nil {
			return raw.(error)
		}

		// Each our values should represent a populated value, so turn that
		// into a row.
		var row table.Row
		for _, v := range rowVals {
			raw := v.Elem().Interface()
			if s, ok := raw.(string); ok {
				raw = wordwrap.WrapString(s, 50)
			}
			row = append(row, raw)
		}
		t.AppendRow(row)
	}

	// If we got an error while iterating, return that and do not render.
	if rows.Err() != nil {
		return rows.Err()
	}

	// Render!
	t.SetStyle(table.StyleRounded)
	t.Render()

	return nil
}

func (c *TestCommand) Flags() *flag.Sets {
	return c.flagSet(flagSetDefault, func(sets *flag.Sets) {
		// Nothing today
	})
}

func (c *TestCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *TestCommand) AutocompleteFlags() complete.Flags {
	return c.Flags().Completions()
}

func (c *TestCommand) Synopsis() string {
	return "Apply a fresh schema to the dev database"
}

func (c *TestCommand) Help() string {
	return formatHelp(`
Usage: squire test [options]

  Run SQL unit tests against the database.

  This command creates a new test container, applies your full schema
  including the test files (ending in "_test.sql") and then runs pgUnit
  against it.

  Currently, Squire only supports pgUnit. We'd like to support pgTAP in
  the future. Squire automatically installs pgUnit into the test database
  so your test SQL files can make use of the functions. pgUnit is installed
  into the "pgunit" schema so you must prefix all pgUnit function calls with
  "pgunit.".

  The test database is always destroyed at the end of the command. If you
  want to debug tests, run "squire reset -include-tests" to reset your
  development database with the test schema. Then you can use "squire console"
  or any other means to debug your database.

` + c.Flags().Help())
}
