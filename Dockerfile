# syntax = docker.io/docker/dockerfile:1.2

# NOTES
#
# - This Dockerfile requires BuildKit. When using `docker build`, set
#   DOCKER_BUILDKIT=1.
#
# - With `RUN --mount=type=bind,rw`, writes to the bind-mounted directory are
#   discarded after the RUN finishes. Ensure that any final build output exists
#   outside of that directory.

FROM docker.io/library/golang:1.15-alpine3.13 AS golang
FROM golang AS server-build

RUN apk add --no-cache \
      build-base \
      gstreamer-dev

RUN \
  --mount=type=bind,target=/mnt/hypcast \
  --mount=type=cache,id=hypcast.go-pkg,target=/go/pkg \
  --mount=type=cache,id=hypcast.go-build,target=/root/.cache/go-build,from=golang,source=/root/.cache/go-build \
  cd /mnt/hypcast && \
  go build -v \
    -trimpath -ldflags='-s -w' \
    -o /hypcast-server \
    ./cmd/hypcast-server

# ---

FROM --platform=$BUILDPLATFORM docker.io/library/node:14-alpine3.12 AS client-build

RUN \
  --mount=type=bind,target=/mnt/hypcast,rw \
  --mount=type=cache,id=hypcast.node_modules,target=/mnt/hypcast/client/node_modules \
  --mount=type=cache,id=hypcast.yarn,target=/usr/local/share/.cache/yarn \
  cd /mnt/hypcast/client && \
  yarn install --production --frozen-lockfile && \
  yarn build && \
  mv build /build

# ---

FROM docker.io/library/alpine:3.13 AS target

RUN apk add --no-cache \
      tini \
      gstreamer \
      gst-plugins-base \
      gst-plugins-good \
      gst-plugins-bad \
      gst-plugins-ugly && \
    mkdir -p \
      /opt/hypcast/bin \
      /opt/hypcast/share/www

COPY --from=server-build /hypcast-server /opt/hypcast/bin/hypcast-server
COPY --from=client-build /build /opt/hypcast/share/www

EXPOSE 9200
ENTRYPOINT [ \
  "/sbin/tini", "--", \
  "/opt/hypcast/bin/hypcast-server", \
  "-assets", "/opt/hypcast/share/www" \
]
