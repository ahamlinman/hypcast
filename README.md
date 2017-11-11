# Hypcast

Hypcast is an interactive web app designed to let you watch live TV from a
[LinuxTV][1]-compatible tuner card in your browser. On the frontend, you select a
channel and video quality. On the backend, a Node app manages the `azap` and
`ffmpeg` programs and produces an [HLS][2] video stream. This lets you bring
your TV experience anywhere in the world for free.

[1]: https://www.linuxtv.org/wiki/index.php/Main_Page
[2]: https://en.wikipedia.org/wiki/HTTP_Live_Streaming

## How it Works

When you first open the Hypcast UI, you will be presented with a list of
channel and video quality options. Once you select a channel and quality, you
can use the start button to start the stream. This will activate your tuner,
tune it to the selected channel, and start encoding the video stream in real
time. Once there is enough video encoded for your browser to download, a video
player will appear and the stream will begin playing. You can select another
channel and quality and press the start button again to tune to a different
channel, or you can press the stop button to stop the stream.

The state of the tuner is synchronized across all Hypcast clients. In other
words, if you are watching on your laptop and visit Hypcast on your phone at
the same time, you will see the same video stream that you are already
watching (though you will probably start at an earlier point in the stream).
If you stop the stream or change the channel on your phone, the change will be
reflected almost immediately on your laptop and vice-versa. Anyone who can
connect to Hypcast has an equal ability to control the tuner. This makes it
easy to change channels from your phone or pick up where you left off on
another device, but it means that you should be careful about who can access
your Hypcast instance.

## Assumptions

Hypcast is currently built with the following assumptions and design choices:

* It is designed for a single tuner per system. It should be possible to
  support multiple tuners on the backend, though this will require some API
  and UI changes.
* It is designed for ATSC tuners (i.e. those that can be controlled with the
  azap utility) only. However, it should be very easy to support the other
  \*zap utilities by working from the existing azap code.
* It is designed for any client to have equal access to control the tuner. You
  should probably not make a Hypcast instance public, as it is too easy to
  abuse if you're not careful.
* It is designed for live TV only. Hypcast generally keeps enough of the
  stream around for you to rewind or pause for a few minutes at a time, but it
  destroys the stream when stopped and does not support any kind of recording
  functionality. There are already far better solutions for this use case,
  like MythTV and Tvheadend.
* The web UI relies on native support for ES2017 features, and thus only runs
  in *very* modern browsers. The latest versions of Chrome, Firefox, Edge, or
  Safari should work. I can guarantee that *no* version of Internet Explorer
  will *ever* work.

## Installation

Hypcast requires node, dvb-apps and ffmpeg with libfdk\_aac support. A
Dockerfile is included in this repository that will nicely roll all of these
dependencies into a Docker image, along with some npm scripts to help build
and run the image. By default, this runs Hypcast as a non-root user and makes
it available on port 9400 (Hypcast's default port) on the host.

## Usage

Hypcast requires two configuration files: `channels.conf` and `profiles.json`.
Examples are provided in the `doc/` directory of this repository.

* `channels.conf` is a list of channels suitable for use with the `azap`
  utility. This can be generated using a utility like `scan` or `w_scan`. For
  more information, see https://www.linuxtv.org/wiki/index.php/Frequency\_scan
* `profiles.json` contains sets of options that will be passed to ffmpeg when
  encoding the stream, such as the video size, x264 preset, audio bitrate,
  etc. Multiple profiles give you the flexibility to select stream parameters
  based on the nature of your connection to the computer where Hypcast is
  running. For example, if you are on a mobile network you can select a
  lower-quality profile to conserve your data. The example profiles should be
  sufficient for a reasonably well-powered desktop computer, but they have
  been determined through trial and error and may not be appropriate for all
  systems.

When running Hypcast in a Docker container, a directory containing these files
should be mounted read-only as `/hypcast/config`. You will also need to give
the container access to your TV tuner devices using Docker's `--device`
option (e.g. `--device=/dev/dvb`).

## Additional Questions

You can view the most up-to-date methods for contacting me at
http://alexhamlin.co.
