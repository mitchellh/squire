package cli

import (
	"strings"
)

// formatHelp should be called around all Help text for commands. This
// applies final formatting rules.
func formatHelp(v string) string {
	return strings.TrimSpace(v)
}
