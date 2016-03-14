import $ from 'jquery';
import machina from 'machina';

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
            let profileList = $('#quality');
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
      },

      loadComplete() {
        if (this.profiles && this.channels) {
          this.transition('connecting');
        }
      },

      error() {
        this.transition('error');
      },
    },

    connecting: {
      _onEnter() {
        console.log('started connecting');
      },
    },

    error: {
      _onEnter() {
        $('.hyp-ui').hide();
        $('.hyp-error').show();
      },

      _onExit() {
        $('.hyp-error').hide();
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
