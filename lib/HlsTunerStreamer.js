import Machina from 'machina';
import FfmpegCommand from 'fluent-ffmpeg';
import Tmp from 'tmp';
import fs from 'fs';
import touch from 'touch';
import path from 'path';

const TunerMachine = Machina.Fsm.extend({
  initialize(tuner) {
    this._tuner = tuner;

    this._tuner.on('lock', (chan) => this.handle('tunerLock', chan));
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
      // We begin by starting up the tuner...
      _onEnter() {
        this._tuner.tune(this._tuneData.channel);
      },

      // ...and we'll start FFmpeg when we have a signal lock
      tunerLock(chan) {
        this.transition('buffering');
      },

      // If anything goes wrong, just give up
      tunerError(err) {
        this.emit('error', err);
        this.transition('inactive');
      },

      tunerStop() {
        this.emit('error', new Error('The tuner unexpectedly stopped'));
        this.handle('stop');
      },

      // If the user quickly tries to change the channel, just go back to the
      // inactive state and start tuning again from there. A clean start.
      tune(data) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      // If we stop from this state, make sure the tuner gets stopped
      stop() {
        this.transition('detuning');
      },
    },

    buffering: {
      // Start by making a temp directory for the encoded files. This will be
      // removed when the streamer is stopped.
      _onEnter() {
        Tmp.dir({ unsafeCleanup: true }, (err, path, clean) => {
          if (err) {
            this.emit('error', err);
            this.transition('detuning');
          }

          this._ffmpegCleanup = clean;
          this.streamPath = path;

          this.handle('tmpPathReady');
        });
      },

      // Once we have a temp directory, create a playlist file so that we can
      // watch it for changes. When we start FFmpeg, it will take a little
      // while to get the HLS stream going. Once the first .ts file is ready
      // for the client, the playlist file will get updated. That's when we
      // make the streamer's state active and notify the client that they can
      // start watching.
      tmpPathReady() {
        this.playlistPath = path.join(this.streamPath, 'stream.m3u8');

        touch(this.playlistPath, (err) => {
          if (err) {
            this.emit('error', err);
            this.transition('debuffering');
          }

          this.handle('playlistReady');
        });
      },

      // Now that our dummy playlist file is ready to go, we watch it for real
      // and start up FFmpeg so that it will eventually get changed.
      playlistReady() {
        let watcher = fs.watch(this.playlistPath)
          .once('change', () => {
            watcher.close();
            this.handle('buffered');
          })
          .on('error', (err) => {
            this.emit('error', err);
            this.transition('debuffering');
          });

        let profile = this._tuneData.profile;

        this._ffmpeg = new FfmpegCommand(this._tuner.device)
          .complexFilter([`scale=-2:ih*min(1\\,${profile.videoHeight}/ih)`])
          .videoCodec('libx264').videoBitrate(profile.videoBitrate)
          .outputOptions([
              '-profile:v main',
              '-tune zerolatency',
              `-bufsize ${profile.videoBufsize}`,
              `-preset ${profile.videoPreset}`,
          ])
          .audioCodec('libfdk_aac').audioBitrate(profile.audioBitrate)
          .outputOptions(`-profile:a ${profile.audioProfile}`)
          .outputOptions(['-f hls', '-hls_wrap 20', '-hls_list_size 20'])
          .on('start', (cmd) => console.log('ffmpeg started:', cmd))
          .on('error', (err) => this.handle('ffmpegError', err))
          .on('end', () => this.handle('ffmpegEnd'))
          .save(this.playlistPath);
      },

      // At this point, the first .ts file is ready and the client can download
      // the playlist
      buffered() {
        this.transition('active');
      },

      ffmpegError(err) {
        this.emit('error', err);
        this.handle('stop');
      },

      ffmpegEnd() {
        this.handle('stop');
      },

      tunerError(err) {
        this.emit('error', err);
        this.handle('stop');
      },

      tunerStop() {
        this.emit('error', new Error('The tuner unexpectedly stopped'));
        this.handle('stop');
      },

      tune(data) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop() {
        this.transition('debuffering');
      },
    },

    active: {
      // If the user changes the channel, just start everything over
      tune(data) {
        // If this console.log line is not here, the stream will immediately
        // stop when the active state is reached. WTF?!?
        console.log('(retuning)');
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop() {
        this.transition('debuffering');
      },

      tunerError(err) {
        this.emit('error', err);
        this.handle('stop');
      },

      tunerStop() {
        this.emit('error', new Error('The tuner unexpectedly stopped'));
        this.handle('stop');
      },

      ffmpegError(err) {
        this.emit('error', err);
        this.handle('stop');
      },

      ffmpegEnd() {
        this.emit('error', new Error('FFmpeg unexpectedly stopped'));
        this.handle('stop');
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

      tune(data) {
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

class HlsTunerStreamer extends TunerMachine {
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

export default HlsTunerStreamer;
