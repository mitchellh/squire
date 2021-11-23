#--------------------------------------------------------------------
# pgq builds pgquarrel
#--------------------------------------------------------------------

FROM alpine:3.15 AS pgq

RUN apk add --no-cache postgresql-dev git cmake make gcc libc-dev

# Clone and build
RUN git clone https://github.com/eulerto/pgquarrel.git /pgquarrel
WORKDIR /pgquarrel
RUN cmake -DCMAKE_PREFIX_PATH=/usr/lib/postgresql14/ .; \
    make; \
    make install;

#--------------------------------------------------------------------
# builder builds the Waypoint binaries
#--------------------------------------------------------------------

FROM golang:1.17-alpine AS builder

RUN apk add --no-cache git gcc libc-dev make

RUN mkdir -p /tmp/squire-mod
COPY go.sum /tmp/squire-mod
COPY go.mod /tmp/squire-mod
WORKDIR /tmp/squire-mod
RUN go mod download

COPY . /tmp/squire-src
WORKDIR /tmp/squire-src

RUN --mount=type=cache,target=/root/.cache/go-build make

#--------------------------------------------------------------------
# final image
#--------------------------------------------------------------------

FROM alpine:3.15

RUN apk add --no-cache postgresql-dev postgresql

# pgquarrel
COPY --from=pgq /usr/local/bin/pgquarrel /usr/local/bin/pgquarrel
COPY --from=pgq /usr/local/lib/libmini.so /usr/local/lib/libmini.so

# squire
COPY --from=builder /tmp/squire-src/bin/squire /usr/local/bin/squire

ENTRYPOINT ["/usr/local/bin/squire"]
