import * as socketio from 'socket.io-client';
import { EventEmitter } from 'events';

import { TuneData } from '../models/TuneData';

/**
 * This class uses Socket.io to synchronize the client with a state machine on
 * the Hypcast server. On each transition of the server's state machine, this
 * client will update its internal state and emit an event. Downstream, we can
 * use these events to do something useful (for example, telling React to
 * re-render the UI).
 *
 * This class also provides functions for the client to modify the server's
 * state. Obviously the server should control whether we are tuning a channel
 * or waiting for video to buffer, but clients control when the tuner starts
 * and stops. On these actions, we actually ask the server to modify *its*
 * state machine, and *its* resulting transitions get reflected on the client.
 * This ensures that everything stays perfectly syncrhonized across all
 * clients.
 *
 * One more small detail: When we connect to the server, it will immediately
 * fire off an event over the socket to transition *us* to its state. That's
 * why nothing explicit happens here to sync us on initialization.
 */

interface Transition {
  toState: string;
  tuneData: TuneData;
}

export default class HypcastController extends EventEmitter {
  private _state: string = 'connecting';
  private _socket: SocketIOClient.Emitter;

  constructor() {
    super();

    this._socket = socketio()
      .on('transition', ({ toState, tuneData }: Transition) => {
        if (tuneData) {
          this.emit('updateTuning', tuneData);
        }

        this.state = toState;
      })
      .on('disconnect', () => {
        this.state = 'connecting';
      });
  }

  get state() {
    return this._state;
  }

  set state(toState: string) {
    this._state = toState;
    this.emit('transition');
  }

  tune(tuneData: TuneData) {
    this._socket.emit('tune', tuneData);
  }

  stop() {
    this._socket.emit('stop');
  }
}
