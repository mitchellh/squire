package cli

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"

	"github.com/mitchellh/squire/internal/pkg/signalcontext"
	"github.com/mitchellh/squire/internal/version"
)

const (
	// cliName is the name of this CLI.
	cliName = "squire"
)

// Main runs the CLI with the given arguments and returns the exit code.
// The arguments SHOULD include argv[0] as the program name.
func Main(args []string) int {
	// Get our version
	vsn := version.Info()

	// NOTE: This is only for running `squire -v` and expecting it to return
	// a version. Any other subcommand will expect `-v` to be around verbose
	// logging rather than printing a version
	if len(args) == 2 && args[1] == "-v" {
		args[1] = "-version"
	}

	// Initialize our logger based on env vars
	args, log, err := logger(args)
	if err != nil {
		panic(err)
	}

	// Build our cancellation context
	ctx, closer := signalcontext.WithInterrupt(context.Background(), log)
	defer closer()

	// Get our base command
	base, commands := Commands(ctx, log)
	defer base.Close()

	// Build the CLI. We use a CLI factory function because to modify the
	// args once you call a func on CLI you need to create a new CLI instance.
	cliFactory := func() *cli.CLI {
		return &cli.CLI{
			Name:                       args[0],
			Args:                       args[1:],
			Version:                    vsn.String(),
			Commands:                   commands,
			Autocomplete:               true,
			AutocompleteNoDefaultFlags: true,
			HelpFunc:                   cli.BasicHelpFunc(cliName),
		}
	}

	// Copy the CLI to check if it is a version call. If so, we modify
	// the args to just be the version subcommand. This ensures that
	// --version behaves by calling `squire version` and we get consistent
	// behavior.
	cli := cliFactory()
	if cli.IsVersion() {
		// We need to reinit because you can't modify fields after calling funcs
		cli = cliFactory()
		cli.Args = []string{"version"}
	}

	// Run the CLI
	exitCode, err := cli.Run()
	if err != nil {
		panic(err)
	}

	return exitCode
}

// commands returns the map of commands that can be used to initialize a CLI.
func Commands(
	ctx context.Context,
	log hclog.Logger,
) (*baseCommand, map[string]cli.CommandFactory) {
	baseCommand := &baseCommand{
		Ctx: ctx,
		Log: log,
	}

	// start building our commands
	commands := map[string]cli.CommandFactory{
		"config": func() (cli.Command, error) {
			return &ConfigCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"diff": func() (cli.Command, error) {
			return &DiffCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"schema": func() (cli.Command, error) {
			return &SchemaCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"console": func() (cli.Command, error) {
			return &ConsoleCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"up": func() (cli.Command, error) {
			return &UpCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"down": func() (cli.Command, error) {
			return &DownCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"reset": func() (cli.Command, error) {
			return &ResetCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"deploy": func() (cli.Command, error) {
			return &DeployCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"test": func() (cli.Command, error) {
			return &TestCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"url": func() (cli.Command, error) {
			return &URLCommand{
				baseCommand: baseCommand,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &VersionCommand{
				baseCommand: baseCommand,
			}, nil
		},
	}

	return baseCommand, commands
}

// logger returns the logger to use for the CLI. Output, level, etc. are
// determined based on environment variables if set.
func logger(args []string) ([]string, hclog.Logger, error) {
	app := args[0]

	// Determine our log level if we have any. First override we check if env var
	level := hclog.NoLevel

	// Process arguments looking for `-v` flags to control the log level.
	// This overrides whatever the env var set.
	var outArgs []string
	for i, arg := range args {
		if len(arg) != 0 && arg[0] != '-' {
			outArgs = append(outArgs, arg)
			continue
		}

		// If we hit a break indicating pass-through flags, we add them all to
		// outArgs and just exit, since we don't want to process any secondary
		//  `-v` flags at this time.
		if arg == "--" {
			outArgs = append(outArgs, args[i:]...)
			break
		}

		switch arg {
		case "-v":
			if level == hclog.NoLevel || level > hclog.Info {
				level = hclog.Info
			}
		case "-vv":
			if level == hclog.NoLevel || level > hclog.Debug {
				level = hclog.Debug
			}
		case "-vvv":
			if level == hclog.NoLevel || level > hclog.Trace {
				level = hclog.Trace
			}
		default:
			outArgs = append(outArgs, arg)
		}
	}

	// Default output is nowhere unless we enable logging.
	var output io.Writer = ioutil.Discard
	color := hclog.ColorOff
	if level != hclog.NoLevel {
		output = os.Stderr
		color = hclog.AutoColor
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   app,
		Level:  level,
		Color:  color,
		Output: output,
	})

	return outArgs, logger, nil
}
