package squire

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/squire/internal/sqlbuild"
)

type SchemaOptions struct {
	// Output is the location where the schema will be written.
	Output io.Writer

	// Tests will render test SQL files too (files ending in _test.sql).
	// TestsOnly will render only test files and requires Tests to be true.
	Tests     bool
	TestsOnly bool
}

// Schema generates the SQL schema from the SQL directory in the attached
// configuration on the Squire instance.
func (s *Squire) Schema(opts *SchemaOptions) error {
	// Determine our directory. We want an absolute directory so we can
	// put together the fs.FS implementation.
	sqlDir, err := filepath.Abs(s.config.SQLDir)
	if err != nil {
		return fmt.Errorf("Error expanding sql directory: %w", err)
	}
	rootDir, rootFile := filepath.Split(sqlDir)
	if len(rootDir) > 0 && rootDir[len(rootDir)-1] == filepath.Separator {
		// Strip the trailining filepath separator.
		rootDir = rootDir[:len(rootDir)-1]
	}

	// Build to our output
	return sqlbuild.Build(&sqlbuild.Config{
		Output:    opts.Output,
		FS:        os.DirFS(rootDir),
		Root:      rootFile,
		Logger:    s.logger.Named("sqlbuild"),
		Tests:     opts.Tests,
		TestsOnly: opts.TestsOnly,
		Metadata: map[string]string{
			"Generation Time": time.Now().Format(time.UnixDate),
		},
	})
}
