/* eslint-disable no-console */

import * as Machina from 'machina';
import * as FfmpegCommand from 'fluent-ffmpeg';
import * as Tmp from 'tmp';
import * as fs from 'fs';
import * as path from 'path';
import { EventEmitter } from 'events';
import { promisify } from 'util';

import { TuneData } from '../models/TuneData';

// Helper function to create a temporary directory using promises. This is
// required because the function has a non-standard callback with multiple
// "success" arguments, which we need both of. As a result, we construct a
// custom object to resolve the promise.
function createTmpDir(): Promise<{ dirPath: string, clean: () => void }> {
  return new Promise((resolve, reject) => {
    Tmp.dir({ unsafeCleanup: true }, (err, dirPath, clean) => {
      if (err) {
        reject(err);
      }

      resolve({ dirPath, clean });
    });
  });
}

// Helper function to create a new file. This simply helps create the playlist
// file before watching it. We require the path to be exclusive as a safety
// measure. Hypcast shouldn't have a problem with this since a new temp dir is
// created for every stream.
async function createNewFile(filePath: string) {
  const fd = await promisify(fs.open)(filePath, 'wx');
  await promisify(fs.close)(fd);
}

interface Tuner extends EventEmitter {
  device: string;

  tune(channel: string): void;
  stop(): void;
}

interface TunerMachine extends EventEmitter {
  _ffmpeg: typeof FfmpegCommand | null;
  _ffmpegCleanup: () => void | null;
  _tuneData: TuneData | null;
  _tuner: Tuner;

  state: string;
  streamPath: string;
  playlistPath: string | null;

  new(tuner: any): TunerMachine;
  handle: (action: string, ...args: any[]) => void;
  transition: (state: string) => void;
  deferUntilTransition: (state: string) => void;
}

// Common error handlers for while the tuner / streamer are initializing or
// active. In all cases, everything is stopped and errors are dumped where
// possible.
const ErrorHandlers = {
  tunerError(this: TunerMachine, err: Error) {
    this.emit('error', err);
    this.handle('stop');
  },

  tunerStop(this: TunerMachine) {
    this.emit('error', new Error('The tuner unexpectedly stopped'));
    this.handle('stop');
  },

  ffmpegError(this: TunerMachine, err: Error, stdout: any, stderr: any) {
    console.error('Dumping FFmpeg output:', stdout, stderr);
    this.emit('error', err);
    this.handle('stop');
  },

  ffmpegEnd(this: TunerMachine, stdout: any, stderr: any) {
    console.error('Dumping FFmpeg output:', stdout, stderr);
    this.emit('error', new Error('FFmpeg unexpectedly stopped'));
    this.handle('stop');
  },
};

const TunerMachine: TunerMachine = Machina.Fsm.extend({
  initialize(this: TunerMachine, tuner: Tuner) {
    this._tuner = tuner;

    this._tuner.on('lock', () => this.handle('tunerLock'));
    this._tuner.on('error', (err: Error) => this.handle('tunerError', err));
    this._tuner.on('stop', () => this.handle('tunerStop'));
  },

  initialState: 'inactive',
  states: {
    inactive: {
      _onEnter(this: TunerMachine) {
        delete this._tuneData;
      },

      // The user wants to stream a given channel
      tune(this: TunerMachine, data: TuneData) {
        this._tuneData = data;
        this.transition('tuning');
      },
    },

    tuning: {
      ...ErrorHandlers,

      // We begin by starting up the tuner...
      _onEnter(this: TunerMachine) {
        if (!this._tuneData) { throw new Error('no tuneData while tuning'); }
        this._tuner.tune(this._tuneData.channel);
      },

      // ...and we'll start FFmpeg when we have a signal lock
      tunerLock(this: TunerMachine) {
        this.transition('buffering');
      },

      // If the user quickly tries to change the channel, just go back to the
      // inactive state and start tuning again from there. A clean start.
      tune(this: TunerMachine) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      // If we stop from this state, make sure the tuner gets stopped
      stop(this: TunerMachine) {
        this.transition('detuning');
      },
    },

    buffering: {
      ...ErrorHandlers,

      async _onEnter(this: TunerMachine) {
        try {
          // Start by making a temp directory for the encoded files. This will
          // be removed when the streamer is stopped.
          const { dirPath, clean } = await createTmpDir();
          this.streamPath = dirPath;
          this._ffmpegCleanup = clean;

          // Once we have a temp directory, create a playlist file so that we
          // can watch it for changes. When we start FFmpeg, it will take a
          // little while to get the HLS stream going. Once the first .ts file
          // is ready for the client, the playlist file will get updated.
          // That's when we make the streamer's state active and notify the
          // client that they can start watching.
          this.playlistPath = path.join(this.streamPath, 'stream.m3u8');
          await createNewFile(this.playlistPath);
        } catch (err) {
          this.emit('error', err);
          this.transition('debuffering');
          return;
        }

        // Now that our dummy playlist file is ready to go, we watch it for
        // real and start up FFmpeg so that it will eventually get changed.
        const watcher = fs.watch(this.playlistPath)
          .once('change', () => {
            watcher.close();
            this.handle('buffered');
          })
          .on('error', (err: Error) => {
            this.emit('error', err);
            this.transition('debuffering');
          });

        if (!this._tuneData) { throw new Error('no tuneData while buffering'); }
        if (!this._tuneData.profile) { throw new Error('tuneData has no profile'); }

        const { profile } = this._tuneData;

        // Let's go through everything that FFmpeg is doing...
        this._ffmpeg = new FfmpegCommand({ source: this._tuner.device, logger: console })
          .inputOptions([
            // Analyze only the first 2 seconds of video to determine the
            // format. This should make the stream start faster, and manual
            // testing has shown this to be generally enough time.
            '-analyzeduration 2000000',
          ])
          .complexFilter([
            // This scales the video down to videoHeight, unless it is already
            // smaller. -2 means that the width will proportionally match.
            `scale=-2:ih*min(1\\,${profile.videoHeight}/ih)`,
            // This will "stretch/squeeze [audio] samples to the given
            // timestamps." The goal is to let audio get back in sync after
            // reading a corrupted stream (e.g. if we're using an antenna and
            // the signal strength is low).
            'aresample=async=1000',
          ])
          .videoCodec('libx264').videoBitrate(profile.videoBitrate)
          .outputOptions([
            // Modern mobile devices should support the H.264 Main Profile.
            // This gives us a small quality boost at the same bitrate.
            '-profile:v main',
            // Naturally, this option optimizes x264 for faster encoding. From
            // what I understand this basically means inserting more I-frames
            // than usual.
            '-tune zerolatency',
            // This will determine how often FFmpeg computes the average
            // bitrate for the stream, to keep it within the specified
            // videoBitrate. Personally, I keep this relatively small (128k for
            // smaller output, 256k for larger output).
            `-bufsize ${profile.videoBufsize}`,
            // This helps set a tradeoff between encoding speed and video
            // quality, based on the videoHeight and how powerful your computer
            // is. Ideally, set this to be as slow as your computer can handle
            // while still encoding the video in realtime. Personally, I can
            // handle "fast" for 240p, "veryfast" for 480p, and need
            // "ultrafast" for anything beyond that.
            `-preset ${profile.videoPreset}`,
            // This "group of pictures" option forces more frequent keyframes,
            // in an attempt to make Hypcast streams start up more quickly.
            '-g 120',
          ])
          // libfdk_aac is the only AAC encoder that supports High-Efficiency
          // AAC. If your playback device supports it, audioProfile == aac_he
          // will noticeably improve quality at low bitrates.
          .audioCodec('libfdk_aac').audioBitrate(profile.audioBitrate)
          .outputOptions(`-profile:a ${profile.audioProfile}`)
          // hls_list_size is a tradeoff between disk usage and how far back in
          // time you can travel. At the hardcoded option of 20 segments, you
          // can typically pause for or rewind a few minutes. This is
          // considered okay, since Hypcast is designed for *live* streaming.
          .outputOptions(['-f hls', '-hls_list_size 20', '-hls_flags delete_segments'])
          .on('start', (cmd: any) => console.log('ffmpeg started:', cmd))
          .on('error', (err: Error, stdout: any, stderr: any) => this.handle('ffmpegError', err, stdout, stderr))
          .on('end', (stdout: any, stderr: any) => this.handle('ffmpegEnd', stdout, stderr))
          .save(this.playlistPath);
      },

      // At this point, the first .ts file is ready and the client can download
      // the playlist
      buffered(this: TunerMachine) {
        this.transition('active');
      },

      tune(this: TunerMachine) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop(this: TunerMachine) {
        this.transition('debuffering');
      },
    },

    active: {
      ...ErrorHandlers,

      // If the user changes the channel, just start everything over
      tune(this: TunerMachine) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop(this: TunerMachine) {
        this.transition('debuffering');
      },
    },

    debuffering: {
      _onEnter(this: TunerMachine) {
        // Kill FFmpeg if it is running (note that it will emit an error on
        // SIGKILL that we need to absorb)
        if (this._ffmpeg) {
          this._ffmpeg.removeAllListeners('error').once('error', () => {});
          this._ffmpeg.kill();
          delete this._ffmpeg;
        }

        // Clean up any encoded files and playlists
        if (this._ffmpegCleanup) {
          this._ffmpegCleanup();
          delete this._ffmpegCleanup;
          delete this.streamPath;
          delete this.playlistPath;
        }

        this.transition('detuning');
      },

      tune(this: TunerMachine) {
        this.deferUntilTransition('inactive');
      },
    },

    detuning: {
      _onEnter(this: TunerMachine) {
        this._tuner.stop();
      },

      tunerStop(this: TunerMachine) {
        this.transition('inactive');
      },

      tune(this: TunerMachine) {
        this.deferUntilTransition('inactive');
      },
    },
  },
});

export default class HlsTunerStreamer extends TunerMachine {
  tune(data: TuneData) {
    this.handle('tune', data);
  }

  stop() {
    this.handle('stop');
  }

  get tuneData() {
    return this._tuneData;
  }
}
