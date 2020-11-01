# Hypcast

Hypcast is a web-based multi-party live television streamer based on
[GStreamer][gstreamer] and [Pion WebRTC][pion], designed to work with TV
tuner hardware on Linux hosts. Start it up, click a channel, and stream live
TV from wherever you are. Connect with another device and share a perfectly
synchronized live stream, even as you change channels.

**Please note…**

- Hypcast is a personal project with a limited feature set that meets my
  specific needs. It is made available to the public in the event that it
  might be useful to others, however I provide no guarantees about
  maintenance, functionality, or future compatibility.
- This is version 2 of Hypcast, a complete rewrite of the original project
  using a radically different implementation and providing a completely new
  user interface. The original version of Hypcast remains available in the
  `v1/main` branch, however it is not maintained.

[gstreamer]: https://gstreamer.freedesktop.org/
[pion]: https://github.com/pion/webrtc

## Setup Guidelines

There are no "official" instructions or support for running Hypcast, but if
you're willing to work through it and try it out here are some general
guidelines.

Hypcast requires a [supported ATSC tuner card][linuxtv-atsc], along with a
[`channels.conf` file][linuxtv-scan] providing tuning information. The
[w_scan2][w_scan2] utility can generate this file. For example, to scan for
over-the-air channels within the United States:

```sh
w_scan2 -f a -c us -X > channels.conf
```

It is recommended that Hypcast be run using the container image published at
`ghcr.io/ahamlinman/hypcast:latest`, with the following configuration:

- TV tuner devices passed through with `--device /dev/dvb`
- `channels.conf` placed at `/etc/hypcast/channels.conf` inside the
  container, e.g. by putting it at this location on the host and passing
  `-v /etc/hypcast:/etc/hypcast:ro`
- Host networking enabled with `--net host`, to allow WebRTC connections to
  the server without NAT traversal (which Hypcast does not support); the
  `-addr` flag can configure the server port if necessary (default `:9200`)

**Hypcast is not designed to be exposed to the Internet!** It is expected to
run on a fast local network, or _perhaps_ over a private VPN. Allowing public
access could present security issues and/or violate laws in your jurisdiction
(be advised that I am not a legal professional, that the suggestion of this
possibility does not constitute legal advice, and that as a user you are
fully responsible for ensuring that your personal usage of Hypcast complies
with relevant local laws).

[linuxtv-atsc]: https://www.linuxtv.org/wiki/index.php/Hardware_device_information
[linuxtv-scan]: https://www.linuxtv.org/wiki/index.php/Frequency_scan
[w_scan2]: https://github.com/stefantalpalaru/w_scan2

## Potential Improvements

- The UI could use some additional work to ensure robustness against
  failures, e.g. automatic reconnection if the server restarts or whatever.
- There are no controls for data rate and encoding quality like version 1 had.
  Ideally these would adjust automatically based on connection quality,
  possibly with some way for clients to request a lower quality to save data.
  I don't expect this to be trivial to implement.
- The system does not support any form of NAT between the server and client,
  including typical container networking implementations. This would require
  configuring a STUN server.
- The UI is currently hardcoded to connect over insecure WebSockets.
- Subtitles are not supported.
