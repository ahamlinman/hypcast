# syntax = docker.io/docker/dockerfile:1.4

# This Dockerfile is designed to support cross-compilation to a variety of
# target architectures with no emulation required at the host level. All FROM
# directives other than the final FROM scratch target include an explicit
# --platform=$BUILDPLATFORM to ensure that this works as expected.

# The Go builder image and Alpine target image should both use the same release
# of Alpine Linux, to ensure there are no incompatibilities in the final image.
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.17 AS base-alpine
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.19-alpine3.17 AS base-golang

# The Node image for the client UI build is based on Alpine for reduced size,
# however the Alpine release does not need to match that of the server build.
FROM --platform=$BUILDPLATFORM docker.io/library/node:18-alpine AS base-node


FROM --platform=$BUILDPLATFORM base-alpine AS server-sysroot
ARG TARGETARCH TARGETVARIANT
COPY build/hypcast-buildenv.sh /hypcast-buildenv.sh
RUN \
  source /hypcast-buildenv.sh && \
  sysroot_init gcc libc-dev glib-dev a52dec-dev libmpeg2-dev opus-dev x264-dev


FROM --platform=$BUILDPLATFORM server-sysroot AS gst-sysroot
ARG TARGETARCH TARGETVARIANT
RUN source /hypcast-buildenv.sh && sysroot_add gstreamer-dev


FROM --platform=$BUILDPLATFORM base-alpine AS gst-build-base
RUN apk add --no-cache bash git clang lld llvm meson glib-dev flex bison
ARG GST_VERSION=1.20.5
RUN git clone -b $GST_VERSION --depth 1 \
  https://gitlab.freedesktop.org/gstreamer/gstreamer.git /tmp/gstreamer
WORKDIR /tmp/gstreamer
COPY build/hypcast-buildenv.sh /hypcast-buildenv.sh
COPY build/gstreamer-build.bash .


FROM --platform=$BUILDPLATFORM gst-build-base AS gst-build
ARG TARGETARCH TARGETVARIANT
RUN \
  --mount=type=bind,from=gst-sysroot,source=/sysroot,target=/sysroot \
  ./gstreamer-build.bash


FROM --platform=$BUILDPLATFORM base-golang AS server-build-base
# Install host tools for cross-compilation and download Go modules, as these are
# usable across all targets.
RUN apk add --no-cache clang lld pkgconf
RUN \
  --mount=type=bind,target=/mnt/hypcast \
  --mount=type=cache,id=hypcast.go-pkg,target=/go/pkg \
  --mount=type=cache,id=hypcast.go-build,target=/root/.cache/go-build \
  cd /mnt/hypcast && go mod download


FROM --platform=$BUILDPLATFORM server-build-base AS server-build
# Build the hypcast-server binary. See hypcast-buildenv.sh for the setup of
# important Go and cgo-related flags.
ARG TARGETARCH TARGETVARIANT
COPY build/hypcast-buildenv.sh /hypcast-buildenv.sh
RUN \
  --mount=type=bind,from=server-sysroot,source=/sysroot,target=/sysroot,rw \
  --mount=type=bind,from=gst-build,source=/gstreamer/usr/local,target=/sysroot/usr/local \
  --mount=type=bind,target=/mnt/hypcast \
  --mount=type=cache,id=hypcast.go-pkg,target=/go/pkg \
  --mount=type=cache,id=hypcast.go-build,target=/root/.cache/go-build \
  cd /mnt/hypcast && \
  source /hypcast-buildenv.sh && \
  go build -v \
    -ldflags=-extld=clang -buildmode=pie \
    -o /hypcast-server \
    ./cmd/hypcast-server


FROM --platform=$BUILDPLATFORM base-node AS client-build
# Build the Hypcast client UI. These assets are the same for all target
# architectures and will only be built once in a multi-platform build.
ENV BUILD_PATH=/build
RUN \
  --mount=type=bind,target=/mnt/hypcast,rw \
  --mount=type=cache,id=hypcast.node_modules,target=/mnt/hypcast/client/node_modules \
  --mount=type=cache,id=hypcast.yarn,target=/usr/local/share/.cache/yarn \
  cd /mnt/hypcast/client && \
  yarn install --frozen-lockfile && \
  yarn build


FROM --platform=$BUILDPLATFORM base-alpine AS sysroot-target
# Bootstrap a distroless-style root filesystem for the final image on the target
# architecture. We can't directly use an Alpine target image since that would
# require running the target architecture's apk binary, which our host might not
# support. We also have to avoid running any build scripts inside of the chroot,
# (e.g. for Busybox symlinks), as they run in the target architecture's shell
# (sysroot_init takes care of this).
ARG TARGETARCH TARGETVARIANT
COPY build/hypcast-buildenv.sh /hypcast-buildenv.sh
RUN \
  source /hypcast-buildenv.sh && \
  sysroot_init tini glib a52dec libmpeg2 opus x264-libs && \
  mkdir -p \
    /sysroot/opt/hypcast/bin \
    /sysroot/opt/hypcast/share/www


FROM scratch AS target

COPY --link --from=sysroot-target /sysroot /
COPY --link --from=gst-build /gstreamer /
COPY --link --from=server-build /hypcast-server /opt/hypcast/bin/hypcast-server
COPY --link --from=client-build /build /opt/hypcast/share/www

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EXPOSE 9200
ENTRYPOINT [ \
  "/sbin/tini", "--", \
  "/opt/hypcast/bin/hypcast-server", \
  "-assets", "/opt/hypcast/share/www" ]
