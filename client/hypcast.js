import $ from 'jquery';
import { findKey } from 'lodash/object';

import HypcastClientController from './controller';

function updateTunerControls({ channel, profile }) {
  $('#channel').val(channel);
  $('#profile').val(findKey(this.profiles, profile));
}

$(() => {
  let controller = new HypcastClientController();

  controller.on('transition', ({ fromState, toState }) => {
    console.debug(`state machine moving from ${fromState} to ${toState}`);
  });

  controller.on('updateTuning', updateTunerControls);
});
