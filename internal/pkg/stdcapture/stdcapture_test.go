package stdcapture

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapture_out(t *testing.T) {
	require := require.New(t)

	var bufout, buferr bytes.Buffer
	require.NoError(Capture(&bufout, &buferr, func() error {
		fmt.Fprint(os.Stdout, "HELLO\n")
		return nil
	}))
	require.Empty(buferr.String())
	require.Equal("HELLO\n", bufout.String())
}
