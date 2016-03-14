import Machina from 'machina';

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

      tune(data) {
        this._tuneData = data;
        this.transition('tuning');
      },
    },

    tuning: {
      _onEnter() {
        this._tuner.tune(this._tuneData.channel);
      },

      tunerLock(chan) {
        this.transition('active');
      },

      tunerError(err) {
        this.emit('error', err);
        this.transition('inactive');
      },

      tune(data) {
        this.deferUntilTransition('inactive');
        this.handle('stop');
      },

      stop() {
        this.transition('detuning');
      },
    },

    active: {
      stop() {
        this.transition('detuning');
      },
    },

    detuning: {
      _onEnter() {
        this._tuner.stop();
      },

      tunerStop() {
        this.transition('inactive');
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
}

export default HlsTunerStreamer;
