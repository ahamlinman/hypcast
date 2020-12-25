FROM --platform=$BUILDPLATFORM docker.io/library/node:14-alpine3.12 AS client-build
WORKDIR /tmp/hypcast/client

COPY client/package.json client/yarn.lock ./
RUN yarn install

COPY client/ ./
RUN yarn build


FROM docker.io/library/golang:1.16-rc-alpine3.12 AS server-build
WORKDIR /tmp/hypcast

RUN apk add --no-cache \
      build-base \
      gstreamer-dev

COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY client/*.go ./client/
COPY --from=client-build /tmp/hypcast/client/build/ ./client/build/

RUN go build -v \
      -trimpath -ldflags="-s -w" \
      -tags embedclient \
      ./cmd/hypcast-server


FROM docker.io/library/alpine:3.12 AS target
WORKDIR /opt/hypcast

RUN apk add --no-cache \
      tini \
      gstreamer \
      gst-plugins-base \
      gst-plugins-good \
      gst-plugins-bad \
      gst-plugins-ugly

COPY --from=server-build /tmp/hypcast/hypcast-server ./hypcast-server

ENTRYPOINT ["/sbin/tini", "--", "/opt/hypcast/hypcast-server"]
