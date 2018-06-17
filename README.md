![Screenshot of the main Hypcast UI](/doc/screenshot.png)

**Hypcast is an interactive web app that lets you take your TV anywhere.**

Specifically, Hypcast combines all of this:

* A TV tuner card in your computer
* Software from the [LinuxTV] project
* [HTTP Live Streaming] technology

…with a simple web-based interface. The upshot is that you can watch television
streams in any modern web browser, from any device that has access to your
Hypcast server. Thanks to real-time communication, the tuner state is
synchronized across all connected devices. Start on your desktop and move to
your tablet, or control your desktop's stream from your phone!

[LinuxTV]: https://www.linuxtv.org/wiki/index.php/Main_Page
[HTTP Live Streaming]: https://en.wikipedia.org/wiki/HTTP_Live_Streaming

## Requirements

* A Linux server, preferably with [Docker] installed
* A [LinuxTV]-compatible tuner card (see "Hardware device information" on the
  linked wiki page)
  - Currently, only ATSC tuning is supported. However, the ATSC code should be
    fairly adaptable to other tuner types (by running `szap`, `tzap`, etc.
    rather than `azap`).
* A `channels.conf` file suitable for use with the `azap` utility. This can be
  generated with a utility like `scan` or `w_scan`. For more information, see
  https://www.linuxtv.org/wiki/index.php/Frequency\_scan. An example file with
  some over-the-air channels for Seattle, WA is in the `doc/` directory.
* A `profiles.json` file defining the set of video qualities you want to make
  available, based on your server's real-time transcoding capability. Many
  computers are not powerful enough to transcode multiple live video streams
  simultaneously, so Hypcast requires that a specific quality be chosen when a
  stream is started. An example file with my personal settings (used with an
  Intel Core i5-3570k at 3.4 GHz) is in the `doc/` directory.

If you choose to run Hypcast without Docker, you'll also need:

* [ffmpeg], with support for libx264 and libfdk\_aac
* [dvb-apps] from LinuxTV

[Docker]: https://www.docker.com/community-edition
[ffmpeg]: https://www.ffmpeg.org/
[dvb-apps]: https://linuxtv.org/wiki/index.php/LinuxTV_dvb-apps

## Setup

These instructions assume that you'll be running Hypcast in a Docker container.
Helper scripts are defined in `package.json` to help facilitate this use. To
use them, you should install the [Yarn] package manager.

(Note that the `package.json` helpers invoke the `docker` CLI using `sudo`, so
that your user is not required to be in the `docker` group. Remember, users in
the `docker` group effectively have passwordless root access to the system
running the Docker daemon!)

0. Place your `channels.conf` and `profiles.json` files under `/etc/hypcast` on
   your server.
0. Run `yarn run docker:build` to create the Hypcast image.
0. Run `yarn run docker:run` to start a Hypcast container. The container will
   have access to the tuner devices under `/dev/dvb` on your server, and will
   automatically restart if it terminates or if your system is rebooted.
0. Go to http://localhost:9400.

[Yarn]: https://yarnpkg.com/en/docs/install

## Caveats and Limitations

* Only one tuner per system is currently supported. (Multi-tuner support is
  possible, but requires both backend and UI changes.)
* As mentioned above, only ATSC tuning is supported. However, this should be
  easily extendable.
* Hypcast is not designed to support any form of time-shifting. While a stream
  can usually be paused for a few minutes at a time, all data for the stream is
  destroyed as soon as it's stopped. There are already far better solutions for
  this use case (e.g. MythTV or Tvheadend).
* The web UI requires a *very* modern browser. The latest versions of Firefox,
  Chrome, Safari, or Edge should generally work. (I can guarantee that no
  version of Internet Explorer will ever work.)

Finally, **you should not expose the Hypcast UI directly to the Internet.**
First, there is no access control, and all connected clients have an equal
ability to control the stream. Second — keeping in mind the major caveat that
_I am not a lawyer, and this is not legal advice_ — this may very well [violate
copyright laws][1] in your country. Hypcast is designed for your private use of
your private tuner.

[1]: https://en.wikipedia.org/wiki/American_Broadcasting_Cos._v._Aereo,_Inc.

## Additional Questions

You can view the most up-to-date methods for contacting me at
http://alexhamlin.co.
