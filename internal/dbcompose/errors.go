package dbcompose

const (
	errDetailNoFile = `
Squire was requested to load a Docker Compose file and to not use any defaults.
No valid Docker Compose file was found in the paths below. To fix this, please
check your settings and also ensure the Compose file exists.

%v
`
	errDetailNoService = `
Could not find the PostgreSQL service in the Compose file! Squire looks for
the database service by looking for a service with the "x-squire" extension
set. Please mark your PostgreSQL service in your compose file as shown below:

services:
  <service name>:
    x-squire:
      service: true
`

	errDetailMultiService = `
Multiple services in your Compose file were marked as PostgreSQL services
for Squire. Squire currently only supports exactly one service marked
as a service.

Conflicting services: %[1]q, %[2]q

To fix this, remove the "x-squire" configuration from one of the services.
You should only have the "x-squire" configuration on the service you want to
use as the database service.
`
	errDetailNoPort = `
Squire needs to know the port to use to communicate with the PostgreSQL
container on the host, but failed to determine that port. The port is detected
by looking for a port forwarding to the target port %d. The target port defaults
to 5432 but can be overridden using the "targetPort" setting in the "x-squire"
config. Note that this port is the port in the container.

You can fix this by introducing a forwarded port in your compose specification:

services:
  <your db service>:
    ports:
      - "<any host port>:5432"

If you want to change the target port:

services:
  <your db service>:
    ports:
      - "<any host port>:1234"
    x-squire:
      targetPort: 1234
`

	errDetailNoDB = `
Squire needs to know the name of the database within PostgreSQL to use.
This is determined in multiple ways: (1) via the x-squire configuration
(2) looking for a POSTGRES_DB environment variable.

For x-squire, set the "db" field (this is preferred):

services:
  %[1]s:
    x-squire:
      db: myapp

Or set the POSTGRES_DB environment variable:

services:
  %[1]s:
    environment:
      - POSTGRES_DB=myapp
`
)
