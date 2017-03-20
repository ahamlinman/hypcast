import $ from 'jquery';
import { assign } from 'lodash/object';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastClientController from './controller';
import HypcastTitle from './ui/HypcastTitle';
import ControllerBar from './ui/ControllerBar';

$(() => {
  let controller = new HypcastClientController();

  controller.on('transition', ({ fromState, toState }) => {
    console.debug(`state machine moving from ${fromState} to ${toState}`);
  });

  setupTitle(controller);
  setupControllerBar(controller);
});

function setupTitle(controller) {
  controller.on('transition', render);

  function render() {
    ReactDOM.render(
      <HypcastTitle state={controller.state} />,
      document.getElementById('hypcast-title')
    );
  }
}

function setupControllerBar(controller) {
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

  function getEnabled() {
    return controller.state !== 'connecting';
  }

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
      <ControllerBar
        enabled={getEnabled()}
        channels={channels}
        profiles={profiles}
        tuneData={tuneData}
        onTuneDataChange={handleTuneDataChange}
        onTune={handleTune}
        onStop={handleStop} />,
      document.getElementById('controller-bar')
    );
  }
}
