// This is the primary schema for Squire configuration.

// The directory where the SQL files are. Within this directory, only
// SQL files in subdirectories formatted "NN-<name>" are read, where NN is
// some two digit number, i.e. "01-schema". Top-level SQL files in this
// directory are ignored.
sql_dir: *"sql" | string

// Dev settings configure the development container. These are purposely
// limited because for more complex configurations, you can use your own
// Docker Compose file.
dev: {
	// The default container image to use if docker compose is NOT being used.
	default_image: *"postgres:13.4" | string
}

// Production determines the settings for the "production" target when
// used with commands such as diff or deploy.
production: *#prodEnv | #prodExec

//-------------------------------------------------------------------
// Type Definitions

// prodEnv reads the production target by environment variable.
#prodEnv: {
	mode: "env"
	env:  *"PGURI" | string
}

// prodExec executes a command to determine the connection URL to the
// database. The script should output to stdout the connection URL as the
// first line. Additional lines are ignored.
#prodExec: {
	mode: "exec"
	command: [...string]
}
