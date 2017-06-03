FROM node:8-slim
MAINTAINER Alex Hamlin

RUN echo 'deb http://www.deb-multimedia.org stable main non-free' >> \
		/etc/apt/sources.list.d/deb-multimedia.list

RUN apt-get update \
		&& apt-get install -y --force-yes deb-multimedia-keyring \
		&& apt-get update \
		&& apt-get install -y --no-install-recommends libfdk-aac1 ffmpeg dvb-apps \
		&& rm -rf /var/lib/apt/lists/*

RUN mkdir -p /hypcast
WORKDIR /hypcast

COPY . /hypcast
RUN ./scripts/docker-internal-build.sh

RUN useradd -r -g video -d /hypcast -s /sbin/nologin hypcast
USER hypcast

ENTRYPOINT exec npm start
EXPOSE 9400
