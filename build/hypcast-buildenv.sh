#!/bin/sh

export CGO_ENABLED=1 GOOS=linux GOARCH="$TARGETARCH"

case $TARGETARCH in
	amd64) export CARCH=x86_64 CABI=musl;;
	arm64) export CARCH=aarch64 CABI=musl;;
	arm)
		if [ -z ${TARGETVARIANT+_} ]; then echo "linux/arm requires /v7"; return 1; fi
		export CARCH=armv7 CABI=musleabihf GOARM=7;;
	*)
		echo "unsupported architecture $TARGETARCH"
		return 1;;
esac

sysroot_add () {
	apk add -p /sysroot --arch $CARCH --no-cache --no-scripts --allow-untrusted "$@"
}

sysroot_init () {
	mkdir -p /sysroot/etc/apk
	cp /etc/apk/repositories /sysroot/etc/apk/
	sysroot_add --initdb "$@"
}

export CC=clang
export CGO_CFLAGS="--target=$CARCH-alpine-linux-$CABI --sysroot /sysroot"
export CGO_LDFLAGS="-v --target=$CARCH-alpine-linux-$CABI --sysroot /sysroot -pie -fuse-ld=lld"
export PKG_CONFIG_SYSROOT_DIR=/sysroot
export PKG_CONFIG_PATH=/sysroot/usr/lib/pkgconfig:/sysroot/usr/local/lib/pkgconfig
