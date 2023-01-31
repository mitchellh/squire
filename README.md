# Squire - PostgreSQL Database Development Tool

**⚠️ PROJECT ARCHIVED ⚠️**. I archived this project after not working on
it for a couple years, mostly because I wasn't actively using the tool
anymore. I used this tool for awhile for various projects and I _really_
enjoyed the workflow! Feel free to take these ideas as your own, I've
MIT licensed everything.

Squire is a tool for developing, testing, and deploying PostgreSQL schemas.
Squire can optionally replace traditional database migration tooling, or
it can augment it by helping to auto-generate migrations.

Development is done with plain SQL files in an opinionated (but flexible)
folder structure. Testing is done by writing SQL functions and can be used
to test anything in the database. And deployment is done by diffing your target
database with your desired schema, proposing SQL steps to get there, and then
applying those changes.

**Squire requires PostgreSQL 13+.** This could be resolved with some minor
work to the source, there isn't anything strictly required in version 13
that we need, I just didn't need earlier versions. Contributions are welcome
to support earlier versions.

## Installation

Download the [latest release](https://github.com/mitchellh/squire/releases)
for your platform and install `squire ` to [your PATH](https://superuser.com/questions/488173/how-can-i-edit-the-path-on-linux). Verify squire is installed:

```
$ squire --version
squire v0.1.1 (ca3a7e2)

Squire dependency information:
✓ psql      (path: /usr/local/bin/psql)
✓ pgquarrel (path: /usr/local/bin/pgquarrel)
```

**Squire has two runtime dependencies that you must manually install:**

  * `psql` - The `squire console` command requires this. If you don't plan
    on using `squire console`, then you don't need to install this. This
    is typically installed with PostgreSQL, so install PostgreSQL for your
    system.

  * [`pgquarrel`](https://github.com/eulerto/pgquarrel) - This is used for
    `squire diff` and `squire deploy`. Unfortunately at the time of writing,
    there aren't many packages for this available, so you may have to manually
    compile and install this for your platform.

## Usage

For new projects, start by creating a `sql` directory with the structure
below. You must put your SQL files that are part of your schema within
directories named `NN-<name>` where NN is a number and "name" is anything
you want. Note: you can change the directory from `sql` using the
configuration documented later.

```
.
└── sql
    ├── 00-schema
    │   └── schema.sql
    ├── 01-tables
    │   └── tables.sql
    ├── 02-functions
    │   ├── account_with_default_org.sql
    │   └── account_with_default_org_test.sql
    └── ...
 ```

Within the `sql/` directory, add or modify the `.sql` files to create
your structure. Files ending in `_test.sql` are only used in the test
database and are there to define unit tests. These SQL files are never
loaded into development or production schemas.

You can bring up a dev database database in Docker:

	$ squire up

At any point you can view your full schema (raw SQL):

	$ squire schema

You can open a `psql` console at any time to run SQL queries or manually
test parts of your SQL schema.

	$ squire console

After making changes to your schema, you can apply those changes
by either resetting your database (deletes all data but almost always
works) or deploying (preserves data but can fail depending on changes):

	$ squire reset

	or

	$ squire deploy

Squire can launch another test-only database in Docker and run your
unit tests. This will not modify your dev database and is safe to run
at any time.

	$ squire test

When you're ready to deploy, you can view a diff between development
and production. Or just run the deploy command, which whill still require
approval prior to deploying. Deployment does not rely on your dev
or test databases so it is safe to do anytime you're ready.

The production database is specified using the `PGURI` environment variable
by default, but this can be changed through the configuration.

	$ squire diff -production
	$ squire deploy -production

You can also deploy specific refs from your Git repository, which
can be used as a rollback mechanism or as a way to spin up an environment
with a specific history. Note that `deploy` always asks for confirmation
so this is safe anytime:

	$ squire deploy -production -ref=2020-03-01

### Migration Toolings vs. Squire

Squire is able to fully deploy your schema by creating a diff from
your dev schema to your production schema and applying the diff to reach
the target state (similar to tools like [Terraform](https://www.terraform.io/)
for infrastructure). Due to this, you don't _need_ migration tooling.

However, if you want to use migration tooling, the dev and test cycle
of Squire can still be very useful, and you can use the `squire diff`
command to create a starting point for your migrations. In this workflow,
you would never call `squire deploy`, but you'd use the other features
of Squire.

## Configuration

Squire requires zero configuration out of the box. You can view the
current configuration at any time by running "squire config". This includes
documentation for the Squire configuration.

### Custom PostgreSQL Container

For development with `squire up`, Squire by default creates a PostgreSQL
container based on the official "postgres" Docker image. This container
can be fully customized using [Docker Compose](https://docs.docker.com/compose/)
by creating a service with the `x-squire` configuration set, as shown below.
Save this to `docker-compose.yml` within your repository root.

```yaml
version: '3'
services:
  db:
    image: "postgres:13.4"
    ports:
      - "1234:5432"
    environment:
      - POSTGRES_DB=app-dev
      - POSTGRES_HOST_AUTH_METHOD=trust
    x-squire: {}
```

When running `squire up`, Squire will first look for a Docker Compose
configuration with a service configured with `x-squire`. If this is found,
Squire will start this service along with all dependent services in the
Docker Compose file.

Additional configurations can be specified on `x-squire` as documented below:

```yaml
version: '3'
services:
  db:
    image: "postgres:13.4"
    ports:
      - "1234:5432"
    environment:
      - POSTGRES_DB=app-dev
      - POSTGRES_HOST_AUTH_METHOD=trust
    x-squire:
      # The name of the database to use within this container.
      db: "app-dev"

      # The port that PostgreSQL is listening on INSIDE THE CONTAINER.
      # You must specify a port forward to the host for this port. Squire
      # uses this to determine what port to connect to by lookin up the
      # port forwarding (in this case "1234").
      targetPort: 5432
```

### Custom Configuration

To specify custom configuration, create a file named `.squire` in any
directory where you would call the `squire` CLI (or, any parent directories).
You may also specify the configuration in JSON format using `.squire.json`.
The Squire configuration format is [Cue](https://cuelang.org/).

Currently, the default configuration (`squire config -default`) is:

```cue
{
	// The directory where the SQL files are. Within this directory, only
	// SQL files in subdirectories formatted "NN-<name>" are read, where NN is
	// some two digit number, i.e. "01-schema". Top-level SQL files in this
	// directory are ignored.
	sql_dir: "sql"

	// Dev settings configure the development container. These are purposely
	// limited because for more complex configurations, you can use your own
	// Docker Compose file.
	dev: {
		// The default container image to use if docker compose is NOT being used.
		default_image: "postgres:13.4"
	}

	// Production determines the settings for the "production" target when
	// used with commands such as diff or deploy.
	production: {
		mode: "env"
		env:  "PGURI"
	}
}
```
