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

export LLVMTARGET="$CARCH-alpine-linux-$CABI"

sysroot_init () {
	mkdir -p /sysroot/etc/apk
	cp /etc/apk/repositories /sysroot/etc/apk/
	apk add -p /sysroot --arch "$CARCH" --initdb --no-cache --no-scripts --allow-untrusted "$@"
}

export CC=clang
export CGO_CFLAGS="--target=$LLVMTARGET --sysroot /sysroot"
export CGO_LDFLAGS="--target=$LLVMTARGET --sysroot /sysroot -pie -fuse-ld=lld"
export PKG_CONFIG_SYSROOT_DIR=/sysroot
export PKG_CONFIG_PATH=/sysroot/usr/lib/pkgconfig:/sysroot/usr/local/lib/pkgconfig
