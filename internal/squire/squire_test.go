package squire

import (
	"github.com/hashicorp/go-hclog"
)

func init() {
	// For tests, use higher default log level
	hclog.L().SetLevel(hclog.Debug)
}
