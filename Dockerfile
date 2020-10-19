FROM golang:1.15-buster AS server-build

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
      libgstreamer1.0-dev \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /var/tmp/hypcast
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go install -v ./cmd/hypcast-server


FROM node:14-buster AS client-build

WORKDIR /var/tmp/hypcast/client
COPY client/ ./
RUN yarn install && yarn build


FROM debian:buster-slim AS target

RUN apt-get update \
  && apt-get install -y \
      tini \
      libgstreamer1.0 \
      gstreamer1.0-plugins-base \
      gstreamer1.0-plugins-good \
      gstreamer1.0-plugins-bad \
      gstreamer1.0-plugins-ugly \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/hypcast
COPY --from=server-build /go/bin/hypcast-server ./hypcast-server
COPY --from=client-build /var/tmp/hypcast/client/build/ ./assets/

ENTRYPOINT [ \
  "/usr/bin/tini", "--", \
  "/opt/hypcast/hypcast-server", \
  "-assets", "/opt/hypcast/assets" \
]
