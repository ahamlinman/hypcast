import $ from 'jquery';
import Hls from 'hls.js';
import machina from 'machina';
import socketio from 'socket.io-client';

export default machina.Fsm.extend({
  initialState: 'loading',
  states: {
    loading: {
      _onEnter() {
	// Retrieve profiles
	$.get('/profiles')
	  .done((profiles) => {
	    this.profiles = profiles;
	    let profileList = $('#profile');
	    for (let name of Object.keys(profiles)) {
	      let info = profiles[name];
	      profileList.append(
		$('<option>').prop('value', name).html(info.description));
	    }
	    this.handle('loadComplete');
	  })
	  .fail((xhr) => {
	    console.error('Profile retrieval failed:', xhr);
	    this.handle('error');
	  });
      },

      loadComplete() {
	if (this.profiles) {
	  this.transition('connecting');
	}
      },

      error: 'error',
    },

    connecting: {
      _onEnter() {
	$('h1').addClass('text-muted');
	$('#tuner *').prop('disabled', true);

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

	  $('#tuner').submit((event) => {
	    event.preventDefault();
	    let options = {
	      profile: this.profiles[$('#profile').val()],
	      channel: $('#channel').val(),
	    };
	    console.debug('tuning with options:', options);
	    this.socket.emit('tune', options);
	  });

	  $('#stop').click((event) => {
	    event.preventDefault();
	    this.socket.emit('stop');
	  });
	}
      },

      _onExit() {
	$('h1').removeClass('text-muted');
	$('#tuner *').prop('disabled', false);
      },
    },

    inactive: {
      _onEnter() {
	$('#tuner #stop').prop('disabled', true);
      },

      _onExit() {
	$('#tuner #stop').prop('disabled', false);
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
