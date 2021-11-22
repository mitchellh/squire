// Package config can load configuration for Squire.
package config

import (
	"cuelang.org/go/cue"
)

// Config is the final configuration structure for Squire.
type Config struct {
	// Root is the root value for the configuration. This might be nil
	// if the configuration was hand-created.
	Root cue.Value `json:"-"`

	SQLDir string `json:"sql_dir"`

	Dev struct {
		DefaultImage string `json:"default_image"`
	}
}
