import machina from 'machina';
import socketio from 'socket.io-client';

/**
 * This FSM actually used to control the entire Hypcast UI. But now that it's
 * been stripped down to this tiny core, its function may not be immediately
 * obvious.
 *
 * This state machine uses Socket.io to synchronize itself with a nearly
 * equivalent state machine on the Hypcast server. On each transition of the
 * server's state machine, this client state machine will transition as well
 * (and emit an event indicating that it did so). Downstream, we can use these
 * transition events to do something useful (for example, telling React to
 * re-render the UI based on the new state).
 *
 * This machine also provides functions for the client to modify the state of
 * the server machine. Obviously the server should control whether we are
 * tuning a channel or waiting for video to buffer, but clients control when
 * the tuner starts and stops. On these actions, we actually ask the server to
 * modify *its* state machine, and *its* resulting transitions get reflected on
 * the client. This ensures that everything stays perfectly syncrhonized across
 * all clients.
 *
 * One additional small detail that isn't fully reflected here: When we connect
 * to the server, it will immediately fire off an event over the socket to
 * transition *us* to its state. That's why nothing explicit happens here to
 * sync us on initialization.
 */

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
