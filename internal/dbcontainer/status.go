package dbcontainer

// Status is the status of the container. This can be used to determine
// if the container is created, running, etc.
type Status struct {
	// ID of the container in Docker
	ID string

	// Name of the container
	Name string

	// State of the container.
	State State
}

// State is the possible states that a container can be in. The
// states come directly from Docker so it is possible that future
// Docker versions or alternate implementations may have states that
// aren't present in the constants below.
//
// The state will always be lowercased even if the Docker API has casing.
type State string

const (
	Created    State = "created"
	Restarting State = "restarting"
	Running    State = "running"
	Paused     State = "paused"
	Exited     State = "exited"
	Dead       State = "dead"

	// NotCreated is a custom state introduced by Squire and set on a Status
	// when the container cannot be found and is assumed to not be created.
	NotCreated State = "not-created"
)
