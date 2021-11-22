package config

import (
	_ "embed"
	"io/ioutil"

	//"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

//go:embed schema.cue
var schema []byte

// New creates a new configuration from the given set of options.
func New(opts ...Option) (*Config, error) {
	var result Config
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	// Build our context and immediately compile our schema
	cuectx := cuecontext.New()
	value := cuectx.CompileBytes(schema)
	if err := value.Validate(); err != nil {
		// If our schema fails validation, its a huge problem and we need to crash.
		panic(err)
	}

	// Load files
	for _, f := range options.Files {
		bs, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		newVal := cuectx.CompileBytes(bs)
		value = value.Unify(newVal)
	}

	// Load strings
	for _, s := range options.Strings {
		newVal := cuectx.CompileString(s)
		value = value.Unify(newVal)
	}

	// Build our codec for decoding
	if err := value.Decode(&result); err != nil {
		return nil, err
	}

	result.Root = value
	return &result, nil
}

// Option is used to configure New.
type Option func(*options)

// FromFile loads from a file.
func FromFile(path string) Option {
	return func(opts *options) {
		opts.Files = append(opts.Files, path)
	}
}

// FromString loads from a string.
func FromString(v string) Option {
	return func(opts *options) {
		opts.Strings = append(opts.Strings, v)
	}
}

type options struct {
	// Files to load
	Files []string

	// Strings are additional configs to load as strings.
	Strings []string
}
