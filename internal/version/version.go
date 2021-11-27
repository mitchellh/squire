package version

import (
	"bytes"
	"fmt"
)

var (
	// The git commit that was compiled. This will be filled in by the compiler.
	GitCommit string

	// Version information. This can be manually set or filled in via
	// compiler flags.
	Version = "0.1.2"

	// This will be initialized by the package.
	info *VersionInfo
)

// Info is a struct that contains all our version information. This
// is more or less the same information (or derived information) from
// the global variables in this package, but having it as a struct lets
// us more easily pass it around, serialize it, etc.
type VersionInfo struct {
	Revision string
	Version  string
}

// Info returns the current binary version info.
func Info() *VersionInfo {
	return info
}

func init() {
	info = &VersionInfo{
		Revision: GitCommit,
		Version:  Version,
	}
}

func (v *VersionInfo) String() string {
	var versionString bytes.Buffer

	if v.Version == "" {
		return "version unknown"
	}

	fmt.Fprintf(&versionString, "%s", v.Version)
	if v.Revision != "" {
		fmt.Fprintf(&versionString, " (%s)", v.Revision)
	}

	return versionString.String()
}
