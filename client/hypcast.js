import $ from 'jquery';
import { assign } from 'lodash/object';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastClientController from './controller';
import HypcastUi from './ui';

$(() => {
  let controller = new HypcastClientController();

  controller.on('transition', ({ fromState, toState }) => {
    console.debug(`state machine moving from ${fromState} to ${toState}`);
  });

  let channels = [];
  let profiles = {};
  let tuneData = {
    channel: '',
    profile: {}
  };

  render();

  // Retrieve profiles
  $.get('/profiles')
    .done((loadedProfiles) => {
      profiles = loadedProfiles;
      controller.profiles = profiles;
      render();
    })
    .fail((xhr) => {
      console.error('Profile retrieval failed:', xhr);
    });

  // Retrieve channels
  $.get('/channels')
    .done((loadedChannels) => {
      channels = loadedChannels;
      controller.channels = channels;
      render();
    })
    .fail((xhr) => {
      console.error('Channel retrieval failed:', xhr);
    });

  controller.on('updateTuning', (update) => {
    tuneData = update;
    render();
  });

  controller.on('transition', render);

  function handleTuneDataChange(update) {
    tuneData = assign({}, tuneData, update);
    render();
  }

  function handleTune() {
    controller.tune(tuneData);
  }

  function handleStop() {
    controller.stop();
  }

  function render() {
    ReactDOM.render(
      <HypcastUi
        state={controller.state}
        channels={channels}
        profiles={profiles}
        tuneData={tuneData}
        onTuneDataChange={handleTuneDataChange}
        onTune={handleTune}
        onStop={handleStop} />,
      document.getElementById('hypcast-app')
    );
  }
});
