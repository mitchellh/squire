package dbcontainer

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	//"github.com/davecgh/go-spew/spew"

	"github.com/mitchellh/squire/internal/dbcompose"
)

func TestUpDown(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	// init config
	cfg, err := dbcompose.New(dbcompose.WithPath("testdata/compose-v2.yml"))
	require.NoError(err)

	// Init
	ctr, err := New(WithCompose(cfg))
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
