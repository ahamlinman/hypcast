import $ from 'jquery';
import HypcastClientController from './controller';

$(() => {
  new HypcastClientController()
    .on('transition', ({ fromState, toState }) => {
      console.debug(`state machine moving from ${fromState} to ${toState}`);
    });
});
