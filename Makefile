GIT_COMMIT=$$(git rev-parse --short HEAD)
GIT_DIRTY=$$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE=$$(git describe --tags --always --match "v*")
GIT_IMPORT="github.com/mitchellh/squire/internal/version"
GOLDFLAGS="-s -w -X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)"
CGO_ENABLED?=0

.PHONY: bin/squire
bin/squire: # bin creates the binaries and installs them
	CGO_ENABLED=$(CGO_ENABLED) go build \
				-ldflags $(GOLDFLAGS) \
				-o ./bin/squire ./cmd/squire
	cp ./bin/squire $(GOPATH)/bin/squire

.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 docker build . -t squire:dev
