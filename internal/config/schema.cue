// This is the primary schema for Squire configuration.

// The directory where the SQL files are.
sql_dir: *"sql" | string

// Dev settings configure the `squire up/down` behaviors for development.
// For more complex configurations, use your own Docker Compose file.
dev: {
	// The default container image to use if docker compose is NOT being used.
	default_image: *"postgres:13.4" | string
}

// Deploy determines how the production database is deployed.
deploy: *#deployEnv | #deployExec

//-------------------------------------------------------------------
// Type Definitions

// deployEnv reads the deploy target by environment variable.
#deployEnv: {
	mode: "env"
	env:  *"PGURI" | string
}

// deployExec executes a command to determine the connection URL to the
// database. The script should output to stdout the connection URL as the
// first line. Additional lines are ignored.
#deployExec: {
	mode: "exec"
	command: [...string]
}
