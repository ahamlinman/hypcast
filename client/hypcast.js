import $ from 'jquery';
import Hls from 'hls.js';
import machina from 'machina';
import socketio from 'socket.io-client';

import entries from 'object.entries';
if (!Object.entries) {
  entries.shim();
}

const HypcastClientController = machina.Fsm.extend({
  initialState: 'loading',
  states: {
    loading: {
      _onEnter() {
        // Retrieve profiles
        $.get('/profiles')
          .done((profiles) => {
            this.profiles = profiles;
            let profileList = $('#profile');
            for (let [name, options] of Object.entries(profiles)) {
              profileList.append(
                $('<option>').prop('value', name).html(options.description));
            }
            this.handle('loadComplete');
          })
          .fail((xhr) => {
            console.error('Profile retrieval failed:', xhr);
            this.handle('error');
          });

        // Retrieve channels
        $.get('/channels')
          .done((channels) => {
            this.channels = channels;
            let channelList = $('#channel');
            for (let channel of channels) {
              channelList.append(
                $('<option>').prop('value', channel).html(channel));
            }
            this.handle('loadComplete');
          })
          .fail((xhr) => {
            console.error('Channel retrieval failed:', xhr);
            this.handle('error');
          });
      },

      loadComplete() {
        if (this.profiles && this.channels) {
          this.transition('connecting');
        }
      },

      error: 'error',
    },

    connecting: {
      _onEnter() {
        $('h1').addClass('text-muted');
        $('#tuner button').prop('disabled', true);

        this.socket = socketio()
          .on('connect', () => {
            console.debug('connected to socket.io server');
          })
          .on('transition', ({ toState }) => {
            this.transition(toState);
          })
          .on('disconnect', () => {
            this.transition('connecting');
          })
          .on('hypcastError', (err) => {
            console.error('hypcast server error:', err);
            this.emit('hypcastError', err);
          });

        $('#tuner').submit((event) => {
          event.preventDefault();
          let profile = this.profiles[$('#profile').val()];
          profile.name = $('#profile').val();

          let options = {
            profile,
            channel: $('#channel').val(),
          };
          console.debug('tuning with options:', options);
          this.socket.emit('tune', options);
        });

        $('#stop').click((event) => {
          event.preventDefault();
          this.socket.emit('stop');
        });
      },

      _onExit() {
        $('h1').removeClass('text-muted');
        $('#tuner button').prop('disabled', false);
      },
    },

    inactive: {
      _onEnter() {
        $('#tuner #stop').hide();
      },

      _onExit() {
        $('#tuner #stop').show();
      },
    },

    tuning: {
      _onEnter() {
        $('h1').addClass('hyp-tuning');
      },

      _onExit() {
        $('h1').removeClass('hyp-tuning');
      },
    },

    buffering: {
      _onEnter() {
        $('h1').addClass('hyp-buffering');
      },

      _onExit() {
        $('h1').removeClass('hyp-buffering');
      },
    },

    active: {
      _onEnter() {
        $('h1').addClass('text-success');

        let video = $('video');
        video.slideDown();

        this._hls = new Hls();
        this._hls.loadSource('/stream/stream.m3u8');
        this._hls.attachMedia(video[0]);
        this._hls.on(Hls.Events.MANIFEST_PARSED, () => video[0].play());
      },

      _onExit() {
        let video = $('video');
        video[0].pause();
        video.slideUp();
        this._hls.detachMedia(video[0]);
        this._hls.destroy();
        delete this._hls;

        $('h1').removeClass('text-success');
      },
    },

    error: {
      _onEnter() {
        $('.hyp-ui').hide();
        $('.hyp-error').show();
      },

      _onExit() {
        setTimeout(() => $('.hyp-error').hide(), 5000);
        $('.hyp-ui').show();
      },
    },
  },
});

$(() => {
  new HypcastClientController()
    .on('transition', ({ fromState, toState }) => {
      console.debug(`state machine moving from ${fromState} to ${toState}`);
    });
});
