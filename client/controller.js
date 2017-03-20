import machina from 'machina';
import socketio from 'socket.io-client';

export default machina.Fsm.extend({
  initialState: 'connecting',

  states: {
    connecting: {
      _onEnter() {
	if (this.socket) {
	  return;
	}

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
      },
    },

    inactive: {},

    tuning: {},

    buffering: {},

    active: {},
  },

  tune(tuneData) {
    this.socket.emit('tune', tuneData);
  },

  stop() {
    this.socket.emit('stop');
  },
});
