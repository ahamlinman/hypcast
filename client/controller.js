import $ from 'jquery';
import Hls from 'hls.js';
import machina from 'machina';
import socketio from 'socket.io-client';

export default machina.Fsm.extend({
  initialState: 'connecting',
  states: {
    connecting: {
      _onEnter() {
	$('h1').addClass('text-muted');

	if (!this.socket) {
	  this.socket = socketio()
	    .on('connect', () => {
	      console.debug('connected to socket.io server');
	    })
	    .on('transition', ({ toState, tuneData }) => {
	      if (tuneData) {
		this.emit('updateTuning', tuneData);
	      }

	      this.transition(toState);
	    })
	    .on('disconnect', () => {
	      this.transition('connecting');
	    })
	    .on('hypcastError', (err) => {
	      console.error('hypcast server error:', err);
	      this.emit('hypcastError', err);
	    });
	}
      },

      _onExit() {
	$('h1').removeClass('text-muted');
      },
    },

    inactive: {},

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

  tune(tuneData) {
    this.socket.emit('tune', tuneData);
  },

  stop() {
    this.socket.emit('stop');
  },
});
