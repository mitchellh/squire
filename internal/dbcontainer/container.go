package dbcontainer

type Container struct {
	*config
}

// New creates a new Container instance to represent a new or existing
// desired container. For new containers, this will not physically start
// the container until Up is called.
func New(opts ...Option) (*Container, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	// Validation
	// TODO

	return &Container{
		config: cfg,
	}, nil
}

// Up starts the container. If it is already running, this does nothing.
func (c *Container) Up() error {
	return nil
}

// Down stops the container and removes any data associated with it.
func (c *Container) Down() error {
	return nil
}
