package squire

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mitchellh/squire/internal/config"
)

func TestTestPGUnit(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	// Build our config
	cfg, err := config.New(config.FromString(
		`sql_dir: "testdata/pgunit"`))
	require.NoError(err)

	// Build squire
	sq, err := New(WithConfig(cfg))
	require.NoError(err)

	// Our callback to verify results
	cb := func(rows *sql.Rows) error {
		cols, err := rows.Columns()
		if err != nil {
			return err
		}

		for _, col := range cols {
			t.Logf("column: %s", col)
		}

		count := 0
		for rows.Next() {
			count++
		}
		t.Logf("result count: %d", count)

		return rows.Err()
	}

	// Run pgunit
	require.NoError(sq.TestPGUnit(ctx, &TestPGUnitOptions{
		Callback: cb,
	}))
}
