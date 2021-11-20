// Package config can load configuration for Squire.
package config

// Config is the final configuration structure for Squire.
type Config struct {
	Dev struct {
		DefaultImage string `json:"default_image"`
	}
}