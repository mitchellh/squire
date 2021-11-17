package clitest

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStreams returns the stdout, stderr capture. This should be called
// at the beginning of all CLI tests to ensure proper capture. This should
// be called AFTER initializing the loggers which may go to real stdout/err.
func TestStreams(t *testing.T) (*bytes.Buffer, *bytes.Buffer, func()) {
	var outBuf, errBuf bytes.Buffer
	outClose := TestCapture(t, &os.Stdout, &outBuf)
	errClose := TestCapture(t, &os.Stderr, &errBuf)

	return &outBuf, &errBuf, func() {
		outClose()
		errClose()
	}
}

// Modify an output stream to write to the given writer instead.
// Example: TestCapture(t, &os.Stdout, &buf)
func TestCapture(t *testing.T, out **os.File, dst io.Writer) func() {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Modify stdout
	old := *out
	*out = w

	// Copy
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		defer r.Close()
		io.Copy(dst, r)
	}()

	return func() {
		// Close the writer end of the pipe
		w.Sync()
		w.Close()

		// Reset stdout
		*out = old

		// Wait for the data copy to complete to avoid a race reading data
		<-doneCh
	}
}

// TestTempWd changes the working directory to an empty temporary
// directory to avoid side effects.
func TestTempWd(t *testing.T) (string, func()) {
	td, err := ioutil.TempDir("", "clitest")
	require.NoError(t, err)

	return td, func() {
		os.RemoveAll(td)
	}
}
