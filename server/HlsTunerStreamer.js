/* eslint-disable no-console */

import Machina from 'machina';
import FfmpegCommand from 'fluent-ffmpeg';
import Tmp from 'tmp';
import fs from 'fs';
import path from 'path';
import { promisify } from 'util';

// Helper function to create a temporary directory using promises. This is
// required because the function has a non-standard callback with multiple
// "success" arguments, which we need both of. As a result, we construct a
// custom object to resolve the promise.
function createTmpDir() {
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
async function createNewFile(filePath) {
  const fd = await promisify(fs.open)(filePath, 'wx');
  await promisify(fs.close)(fd);
}

// Common error handlers for while the tuner / streamer are initializing or
// active. In all cases, everything is stopped and errors are dumped where
// possible.
const ErrorHandlers = {
  tunerError(err) {
    this.emit('error', err);
    this.handle('stop');
  },

  tunerStop() {
    this.emit('error', new Error('The tuner unexpectedly stopped'));
    this.handle('stop');
  },

  ffmpegError(err, stdout, stderr) {
    console.error('Dumping FFmpeg output:', stdout, stderr);
    this.emit('error', err);
    this.handle('stop');
  },

  ffmpegEnd(stdout, stderr) {
    console.error('Dumping FFmpeg output:', stdout, stderr);
    this.emit('error', new Error('FFmpeg unexpectedly stopped'));
    this.handle('stop');
  },
};

const TunerMachine = Machina.Fsm.extend({
  initialize(tuner) {
    this._tuner = tuner;

    this._tuner.on('lock', () => this.handle('tunerLock'));
    this._tuner.on('error', (err) => this.handle('tunerError', err));
    this._tuner.on('stop', () => this.handle('tunerStop'));
  },

  initialState: 'inactive',
  states: {
    inactive: {
      _onEnter() {
        delete this._tuneData;
      },

      // The user wants to stream a given channel
      tune(data) {
        this._tuneData = data;
        this.transition('tuning');
      },
    },

    tuning: {
      ...ErrorHandlers,

      // We begin by starting up the tuner...
      _onEnter() {
        this._tuner.tune(this._tuneData.channel);
      },

      // ...and we'll start FFmpeg when we have a signal lock
      tunerLock() {
        this.transition('buffering');
      },

      // If the user quickly tries to change the channel, just go back to the
      // inactive state and start tuning again from there. A clean start.
      tune() {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      // If we stop from this state, make sure the tuner gets stopped
      stop() {
        this.transition('detuning');
      },
    },

    buffering: {
      ...ErrorHandlers,

      async _onEnter() {
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
        }

        // Now that our dummy playlist file is ready to go, we watch it for
        // real and start up FFmpeg so that it will eventually get changed.
        const watcher = fs.watch(this.playlistPath)
          .once('change', () => {
            watcher.close();
            this.handle('buffered');
          })
          .on('error', (err) => {
            this.emit('error', err);
            this.transition('debuffering');
          });

        const { profile } = this._tuneData;

        this._ffmpeg = new FfmpegCommand({ source: this._tuner.device, logger: console })
          .complexFilter([
            `scale=-2:ih*min(1\\,${profile.videoHeight}/ih)`,
            'aresample=async=1000',
          ])
          .videoCodec('libx264').videoBitrate(profile.videoBitrate)
          .outputOptions([
            '-profile:v main',
            '-tune zerolatency',
            `-bufsize ${profile.videoBufsize}`,
            `-preset ${profile.videoPreset}`,
          ])
          .audioCodec('libfdk_aac').audioBitrate(profile.audioBitrate)
          .outputOptions(`-profile:a ${profile.audioProfile}`)
          .outputOptions(['-f hls', '-hls_list_size 20', '-hls_flags delete_segments'])
          .on('start', (cmd) => console.log('ffmpeg started:', cmd))
          .on('error', (err, stdout, stderr) => this.handle('ffmpegError', err, stdout, stderr))
          .on('end', (stdout, stderr) => this.handle('ffmpegEnd', stdout, stderr))
          .save(this.playlistPath);
      },

      // At this point, the first .ts file is ready and the client can download
      // the playlist
      buffered() {
        this.transition('active');
      },

      tune() {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop() {
        this.transition('debuffering');
      },
    },

    active: {
      ...ErrorHandlers,

      // If the user changes the channel, just start everything over
      tune() {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop() {
        this.transition('debuffering');
      },
    },

    debuffering: {
      _onEnter() {
        // Kill FFmpeg if it is running
        if (this._ffmpeg) {
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

      tune() {
        this.deferUntilTransition('inactive');
      },
    },

    detuning: {
      _onEnter() {
        this._tuner.stop();
      },

      tunerStop() {
        this.transition('inactive');
      },

      tune() {
        this.deferUntilTransition('inactive');
      },
    },
  },
});

export default class HlsTunerStreamer extends TunerMachine {
  tune(data) {
    this.handle('tune', data);
  }

  stop() {
    this.handle('stop');
  }

  get tuneData() {
    return this._tuneData;
  }
}
