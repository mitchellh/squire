package dbcontainer

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	//"github.com/davecgh/go-spew/spew"
)

func TestUpDown(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	// Init
	ctr, err := New(
		WithComposeFile("testdata/compose-v2.yml"),
		WithService("postgres"),
	)
	require.NoError(err)

	// Launch, ensure we come back down
	defer func() {
		require.NoError(ctr.Down(ctx))
	}()
	require.NoError(ctr.Up(ctx))

	// Connect
	db, err := sql.Open("postgres", ctr.ConnURI())
	require.NoError(err)
	defer db.Close()
}
