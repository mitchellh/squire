package config

import (
	"os"
	"path/filepath"
)

// Filename is the default filename for the Squire configuration.
// This is assumed to be in "cue" format but a suffix can also be given
// to explicitly make it "cue" or "json".
const Filename = ".squire"

// FindPath looks for our configuration file starting at "start" and
// traversing parent directories until it is found. If it is found, the
// path is returned. If it is not found, an empty string is returned.
// Error will be non-nil only if an error occurred.
//
// If start is empty, start will be the current working directory. If
// filename is empty, it will default to the Filename constant.
func FindPath(start, filename string) (string, error) {
	var err error
	if start == "" {
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	if filename == "" {
		filename = Filename
	}

	for {
		path := filepath.Join(start, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		next := filepath.Dir(start)
		if next == start {
			return "", nil
		}

		start = next
	}
}
