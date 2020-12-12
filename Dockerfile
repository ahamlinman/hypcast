FROM golang:1.15-alpine3.12 AS server-build
WORKDIR /var/tmp/hypcast

RUN apk add --no-cache \
      build-base \
      gstreamer-dev

COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go install -v -ldflags="-s -w" -trimpath ./cmd/hypcast-server


FROM node:14-alpine3.12 AS client-build
WORKDIR /var/tmp/hypcast/client

COPY client/package.json client/yarn.lock ./
RUN yarn install

COPY client/ ./
RUN yarn build


FROM alpine:3.12 AS target
WORKDIR /opt/hypcast

RUN apk add --no-cache \
      tini \
      gstreamer \
      gst-plugins-base \
      gst-plugins-good \
      gst-plugins-bad \
      gst-plugins-ugly

COPY --from=server-build /go/bin/hypcast-server ./hypcast-server
COPY --from=client-build /var/tmp/hypcast/client/build/ ./assets/

ENTRYPOINT [ \
  "/sbin/tini", "--", \
  "/opt/hypcast/hypcast-server", \
  "-assets", "/opt/hypcast/assets" \
]
