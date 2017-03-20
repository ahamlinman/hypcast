import $ from 'jquery';
import Hls from 'hls.js';
import machina from 'machina';
import socketio from 'socket.io-client';

export default machina.Fsm.extend({
  initialState: 'connecting',
  states: {
    connecting: {
      _onEnter() {
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
	    });
	}
      },
    },

    inactive: {},

    tuning: {},

    buffering: {},

    active: {
      _onEnter() {
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
