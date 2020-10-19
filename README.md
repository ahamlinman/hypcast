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

- The UI could use some additional work to ensure robustness against
  failures, e.g. automatic reconnection if the server restarts or whatever.
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
- Other limitations carried over from v1 include: no support for subtitles,
  one tuner only, ATSC only. Subtitle support should absolutely be
  implemented. The others would be nice to have.
