package sqlbuild

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-hclog"
)

// Config is the configuration for Build.
//
// Normally I'd use the options pattern but since this is
// an internal package, I prefer to start with a struct cause
// its less verbose to write.
type Config struct {
	// Output where the final SQL file is written.
	Output io.Writer

	// FS is the filesystem to read from and Root is the root
	// to begin walking to find all SQL files.
	FS   fs.FS
	Root string

	// Logger
	Logger hclog.Logger
}

// Build builds the SQL file for the given options. This
// typically walks a directory in lexicographic order, looking for
// files or directories prefixed with "NN-" where NN is numeric. Within
// the directories, files do NOT have to be prefixed.
func Build(cfg *Config) error {
	if cfg.FS == nil {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		cfg.FS = os.DirFS(wd)
	}
	if cfg.Root == "" {
		cfg.Root = "sql"
	}
	if cfg.Logger == nil {
		// TODO: null logger default
		cfg.Logger = hclog.L()
	}

	_, err := fs.Stat(cfg.FS, cfg.Root)
	if err != nil {
		panic(err)
	}

	// Shorthand cause we log a lot
	L := cfg.Logger
	L.Info("building SQL", "root", cfg.Root)

	return fs.WalkDir(cfg.FS, cfg.Root,
		func(p string, d fs.DirEntry, err error) error {
			log := L.With("path", p)
			log.Trace("walking")

			// If we had an error looking at this path, exit. We do this
			// first because this will be called twice for directories
			// (according to fs docs) so we can skip a directory that
			// we want to ignore.
			if err != nil {
				log.Warn("error during walk", "err", err)
				return err
			}

			// Split the path. We are an immediate child if our parent
			// directory is our root. This is important because we only
			// check the format of files/directories if it is an
			// immediate child.
			dir, file := filepath.Split(p)
			dir = filepath.Clean(dir)
			child := dir == cfg.Root
			log.Trace("dir and file split", "dir", dir, "file", file)

			// If we are a child, let's verify we care about this.
			if child && !reNumPrefix.MatchString(file) {
				log.Trace("ignoring non-prefixed path")

				// If it is a directory, we skip it.
				if d.IsDir() {
					return fs.SkipDir
				}

				// If it is a file, everything is fine.
				return nil
			}

			// We aren't a child OR we know we match. If we're
			// a directory, we do nothing. If we're a file, we want
			// to read and append the file contents.
			if d.IsDir() {
				return nil
			}

			// If the extension isn't SQL, ignore.
			if strings.ToLower(filepath.Ext(file)) != ".sql" {
				log.Trace("ignoring non-SQL file")
				return nil
			}

			// SQL file, read and append it to our writer.
			f, err := cfg.FS.Open(p)
			if err != nil {
				log.Warn("error reading file", "err", err)
				return err
			}
			defer f.Close()

			// Write our filename so its easier to find merged content.
			if _, err := fmt.Fprintf(cfg.Output, flowerBox, p); err != nil {
				log.Warn("error writing file header", "err", err)
				return err
			}

			// Append
			if _, err := io.Copy(cfg.Output, f); err != nil {
				log.Warn("error copying file", "err", err)
				return err
			}

			log.Trace("added to output")
			return nil
		})
}

var reNumPrefix = regexp.MustCompile(`^\d\d-`)

const flowerBox = `
---------------------------------------------------------------------
-- File: %s
---------------------------------------------------------------------
`
