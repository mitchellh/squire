package dbcompose

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/compose-spec/compose-go/types"
	"github.com/mitchellh/mapstructure"
)

// composeExtension is the structure of the extension fields in the
// Docker Compose file that can be used to manually specify information
// about the database.
type composeExtension struct {
	DB         string `mapstructure:"db"`
	TargetPort uint32 `mapstructure:"targetPort"`
}

// parseExtension parses our extension configuration from the given
// compose service specification. You can test whether a field is set
// by comparing to the zero value.
func parseExtension(svc *types.ServiceConfig) (composeExtension, error) {
	var result composeExtension

	if len(svc.Extensions) == 0 {
		return result, nil
	}

	raw, ok := svc.Extensions[extName]
	if !ok || raw == nil {
		return result, nil
	}

	if err := mapstructure.WeakDecode(raw, &result); err != nil {
		return result, errors.WithDetailf(
			errors.Newf(
				"failed to decode x-squire for service %q: %w",
				svc.Name,
				err,
			),
			strings.TrimSpace(errDetailParseExt),
			svc.Name,
		)
	}

	return result, nil
}

const (
	// extName is the name of the extension for squire fields in the compose
	// spec.
	extName = "x-squire"
)

const (
	errDetailParseExt = `
Failed to parse the "x-squire" extension information in your Docker service
%q. This is most frequently due to invalid field names or field types. Please
review the "x-squire" configuration for your service %[1]q and compare this to
the documentation.
`
)
