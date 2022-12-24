#!/usr/bin/env bash
set -eu

info () { printf '\033[35m[gstreamer-build.bash] \033[34m%s\033[0m\n' "$*"; }

source /hypcast-buildenv.sh
case $CARCH in
	armv7) export MESONCPU=arm;;
	    *) export MESONCPU=$CARCH;;
esac
export LLVMTARGET="$CARCH-alpine-linux-$CABI"
info "Building for LLVM target $LLVMTARGET (Meson CPU $MESONCPU)"

info "Initializing sysroot"
mksysroot \
	gcc \
	libc-dev \
	gstreamer-dev \
	a52dec-dev \
	libmpeg2-dev \
	opus-dev \
	x264-dev

# See https://mesonbuild.com/Cross-compilation.html
info "Writing Meson --cross-file"
cat > meson-cross.txt <<EOF
[host_machine]
system = 'linux'
cpu_family = '$MESONCPU'
cpu = '$MESONCPU'
endian = 'little'

[binaries]
pkgconfig = 'pkg-config'
c = 'clang'
c_ld = 'lld'
cpp = 'clang++'
cpp_ld = 'lld'
strip = 'llvm-strip'

[properties]
sys_root = '/sysroot'

[constants]
llvm_args = ['--target=$LLVMTARGET', '--sysroot', '/sysroot']

[built-in options]
c_args = llvm_args
c_link_args = llvm_args
cpp_args = llvm_args
cpp_link_args = llvm_args
pkg_config_path = ['/sysroot/usr/lib/pkgconfig']
EOF

info "Starting Meson setup"
meson setup \
  --cross-file=meson-cross.txt \
  -Db_staticpic=true \
  -Db_pie=true \
  -Dauto_features=disabled \
  -Dgpl=enabled \
  -Dbase=enabled \
  -Dgst-plugins-base:app=enabled \
  -Dgst-plugins-base:audioconvert=enabled \
  -Dgst-plugins-base:audioresample=enabled \
  -Dgst-plugins-base:opus=enabled \
  -Dgst-plugins-base:videorate=enabled \
  -Dgst-plugins-base:videoscale=enabled \
  -Dgood=enabled \
  -Dgst-plugins-good:deinterlace=enabled \
  -Dbad=enabled \
  -Dgst-plugins-bad:dvb=enabled \
  -Dgst-plugins-bad:mpegtsdemux=enabled \
  -Dugly=enabled \
  -Dgst-plugins-ugly:a52dec=enabled \
  -Dgst-plugins-ugly:mpeg2dec=enabled \
  -Dgst-plugins-ugly:x264=enabled \
  output

info "Starting Meson build"
meson compile -C output

info "Installing GStreamer to /gstreamer"
meson install --destdir /gstreamer --strip -C output

info "GStreamer build complete"
