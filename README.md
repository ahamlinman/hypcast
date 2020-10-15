# Hypcast v2

Hypcast is a free, open-source, web-based live television streamer, designed
to work with TV tuner hardware on Linux hosts.

This is version 2 of Hypcast, which is in development. Unlike version 1,
which relies on HTTP Live Streaming via the `azap` and `ffmpeg` command-line
tools, version 2 uses [GStreamer][gstreamer] and [Pion WebRTC][pion] to
provide a truly live streaming experience.

[gstreamer]: https://gstreamer.freedesktop.org/
[pion]: https://github.com/pion/webrtc

As of 2020-10-14, version 2 is fully capable of controlling a TV tuner card
and streaming live video. It provides significantly faster startup and
channel switches than version 1, which is the primary issue I hoped to solve
with the new implementation. Numerous areas for improvement remain:

- The UI needs work, both visually and in terms of robustness (e.g. automatic
  reconnection if the server restarts or whatever).
- Video quality is noticeably worse than in version 1. Switching from VP8
  (libvpx) back to H.264 (x264) will probably make a difference, though for
  some reason I've had trouble getting H.264 video to work.
- There are no controls for data rate and encoding quality like version 1 had.
  Ideally these would adjust automatically based on connection quality,
  possibly with some way for clients to request a lower quality to save data.
  I don't expect this to be trivial to implement.
- The system can only be used over local networks. First, the client is
  hardcoded to use insecure WebSockets. Second, we do not implement STUN,
  TURN, or trickle ICE between peers. As long as the system works over a VPN
  for remote access, this is a very low-priority issue for me. Neither
  version of Hypcast has ever been designed to support public access (no
  authorization mechanism, internal state sometimes exposed to clients,
  etc.).

Hypcast version 2 maintains the caveats and limitations of version 1: one
tuner only, ATSC only, no pausing or time shifting, modern browser required,
don't expose it to the Internet.
