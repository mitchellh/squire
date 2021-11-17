package cli

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/squire/internal/cli/clitest"
)

func testCLI(t *testing.T) (*baseCommand, *bytes.Buffer, *bytes.Buffer, func()) {
	// Create the base first so the logger points to the right place.
	base := testBase(t)

	// Setup our streams
	out, err, closeStream := clitest.TestStreams(t)

	// Move our working directory
	_, closeWd := clitest.TestTempWd(t)

	// Create a once so we only close once
	var once sync.Once

	return base, out, err, func() {
		once.Do(func() {
			closeStream()
			closeWd()
		})
	}
}

func testBase(t *testing.T) *baseCommand {
	log := hclog.L()
	log.SetLevel(hclog.Debug)

	return &baseCommand{
		Ctx: context.Background(),
		Log: log,
	}
}
