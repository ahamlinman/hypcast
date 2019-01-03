FROM node:10-stretch-slim AS base

COPY ./build/tini /bin/tini
ENTRYPOINT ["/bin/tini", "--"]

RUN echo 'deb http://www.deb-multimedia.org stretch main non-free' >> \
		/etc/apt/sources.list.d/deb-multimedia.list

COPY ./build/deb-multimedia-keyring_2016.8.1_all.deb /tmp
RUN dpkg -i /tmp/deb-multimedia-keyring_2016.8.1_all.deb

RUN apt-get update \
		&& apt-get install -y --no-install-recommends libfdk-aac1 ffmpeg dvb-apps \
		&& rm -rf /var/lib/apt/lists/*

# TODO: Hypcast runs in my Ubuntu installation, but not my Arch installation.
# On the Arch system, the group setup and ownership of the dvb device is
# different from what the Debian environment in the container expects. I'm
# switching back to root in the short term, but want to find another long-term
# way to fix this (dynamically creating the user when the container is run?).

# RUN useradd -r -G root,video -d /hypcast -s /sbin/nologin hypcast
# USER hypcast


FROM base AS builder

COPY . /hypcast
RUN /hypcast/build/docker-internal-build.sh


FROM base AS dist
LABEL maintainer="Alex Hamlin <alex@alexhamlin.co>"

EXPOSE 9400
WORKDIR /hypcast
CMD ["node", "./dist/server/index.js"]

COPY --from=builder /hypcast/node_modules /hypcast/node_modules
COPY --from=builder /hypcast/dist /hypcast/dist
