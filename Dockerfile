FROM node:9.9.0-slim
MAINTAINER Alex Hamlin <alex@alexhamlin.co>

RUN echo 'deb http://www.deb-multimedia.org jessie main non-free' >> \
		/etc/apt/sources.list.d/deb-multimedia.list

COPY ./build/deb-multimedia-keyring_2016.8.1_all.deb /tmp
RUN dpkg -i /tmp/deb-multimedia-keyring_2016.8.1_all.deb

RUN apt-get update \
		&& apt-get install -y --no-install-recommends libfdk-aac1 ffmpeg dvb-apps \
		&& rm -rf /var/lib/apt/lists/*

RUN mkdir -p /hypcast
WORKDIR /hypcast

COPY . /hypcast
RUN ./build/docker-internal-build.sh

# For some reason, Docker changes the group of my dvb devices to "root" inside
# the container (it's "video" outside the container). I honestly think this
# might be a regression in Docker. In any case, it would be nice to remove the
# "root" group from this user in the future.
RUN useradd -r -G root,video -d /hypcast -s /sbin/nologin hypcast
USER hypcast

ENTRYPOINT exec npm start
EXPOSE 9400
