package dbcontainer

import (
	dockercommand "github.com/docker/cli/cli/command"
	dockercliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
)

// dockerCli initializes the Docker CLI package. This can be used to
// then get access to the Docker API client and others.
//
// Note we don't use the docker client lib directly because (1) compose
// requires the full CLI anyways and (2) we expect to run on a dev machine
// which we expect has Docker.
func dockerCli() (dockercommand.Cli, error) {
	// Initialize the docker CLI 💀. The compose library needs this to operate.
	cli, err := dockercommand.NewDockerCli()
	if err != nil {
		return nil, err
	}
	if err := cli.Initialize(dockercliflags.NewClientOptions()); err != nil {
		return nil, err
	}

	return cli, nil
}

// dockerClient returns a configured Docker API client.
//
// Note this reinitializes the full CLI currently so if you already have
// the CLI you can just use that.
func dockerClient() (client.APIClient, error) {
	cli, err := dockerCli()
	if err != nil {
		return nil, err
	}

	return cli.Client(), nil
}
