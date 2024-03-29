// stdcapture has helpers for changing Stdout and Stderr temporarily
// and capturing it. This is useful for software that doesn't have
// configurable output streams.
package stdcapture

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// Capture captures the stdout and stderr for the duration of the given
// function and writes it to dstout and dsterr respectively.
func Capture(dstout, dsterr io.Writer, f func() error) error {
	oldout := os.Stdout
	olderr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		return err
	}

	errR, errW, err := os.Pipe()
	if err != nil {
		return err
	}

	// We need to use a waitgroup to wait for the copies to be done
	// to ensure that we have all the data. This makes it so that callers
	// don't have to worry about this.
	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		io.Copy(dstout, outR)
	}()
	go func() {
		defer wg.Done()
		io.Copy(dsterr, errR)
	}()

	// Replace stdout/stderr
	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		// Close is VERY important, cause it ensures that our copy
		// goroutines above will eventually exit
		outW.Close()
		errW.Close()

		os.Stdout = oldout
		os.Stderr = olderr
	}()

	return f()
}

// SuccessOnly captures stdout/stderr but also outputs it back to the
// normal stdout/stderr streams IF there is an error.
func SuccessOnly(dstout, dsterr io.Writer, f func() error) error {
	// Write to both dst and a buffer
	var bufout, buferr bytes.Buffer
	dstout = io.MultiWriter(dstout, &bufout)
	dsterr = io.MultiWriter(dsterr, &buferr)

	// Capture
	err := Capture(&bufout, &buferr, f)

	// On error, just copy it over to stdout/stderr
	if err != nil {
		io.Copy(os.Stdout, &bufout)
		io.Copy(os.Stderr, &buferr)
		return err
	}

	// No error, just return
	return nil
}
