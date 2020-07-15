# Hypcast

Hypcast is a free, open-source, web-based live television streamer.

The first version of Hypcast implements this in a Node.js backend with LinuxTV
dvb-apps, ffmpeg, and HTTP Live Streaming.

This new in-development version of Hypcast will attempt to implement this
functionality in a Go backend with GStreamer and Pion WebRTC. Prototypes of
this strategy based on https://github.com/pion/rtwatch have been promising,
providing fast tuning and a smooth live viewing and listening experience
(GStreamer's WebRTC plugin also works, but the audio is choppy for some
reason).
