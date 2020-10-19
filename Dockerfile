FROM golang:1.15-alpine AS server-build

RUN apk add --no-cache \
      build-base \
      gstreamer-dev

WORKDIR /var/tmp/hypcast
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN go install -v -ldflags="-s -w" ./cmd/hypcast-server


FROM node:14-alpine AS client-build

WORKDIR /var/tmp/hypcast/client
COPY client/ ./

RUN yarn install && yarn build


FROM alpine:3.12 AS target

RUN apk add --no-cache \
      tini \
      gstreamer \
      gst-plugins-base \
      gst-plugins-good \
      gst-plugins-bad \
      gst-plugins-ugly

WORKDIR /opt/hypcast
COPY --from=server-build /go/bin/hypcast-server ./hypcast-server
COPY --from=client-build /var/tmp/hypcast/client/build/ ./assets/

ENTRYPOINT [ \
  "/sbin/tini", "--", \
  "/opt/hypcast/hypcast-server", \
  "-assets", "/opt/hypcast/assets" \
]
