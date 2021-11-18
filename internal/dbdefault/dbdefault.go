// Package dbdefault contains the default configurations for zero config
// Squire usage. This is extracted into its own package so we don't pollute
// the other packages with these kinds of opinions.
package dbdefault

import (
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/types"
)

// Project gets the default Compose project.
func Project() *types.Project {
	wd, err := os.Getwd()
	if err != nil {
		// We don't currently support environments where we don't have
		// a working directory.
		panic(err)
	}

	return &types.Project{
		// Name is our working directory plus a default suffix. We add the
		// suffix because if a user later adds their own docker-compose then
		// this will conflict. We want to be able to detect a default project
		// still running.
		Name: filepath.Base(wd) + "-default",

		WorkingDir: wd,
		Services: []types.ServiceConfig{
			{
				Name:  "postgres",
				Image: "postgres:13.4",
				Ports: []types.ServicePortConfig{
					{
						Target:    5432,
						Published: 5432,
						Protocol:  "tcp",
					},
				},
				Environment: types.NewMappingWithEquals([]string{
					"POSTGRES_DB=squire",
					"POSTGRES_HOST_AUTH_METHOD=trust",
				}),

				Extensions: map[string]interface{}{
					"x-squire": map[string]interface{}{},
				},
			},
		},
	}
}
