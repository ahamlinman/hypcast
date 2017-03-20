import $ from 'jquery';
import { findKey } from 'lodash/object';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastClientController from './controller';
import ChannelSelector from './ui/ChannelSelector';
import ProfileSelector from './ui/ProfileSelector';

$(() => {
  let controller = new HypcastClientController();

  controller.on('transition', ({ fromState, toState }) => {
    console.debug(`state machine moving from ${fromState} to ${toState}`);
  });

  setupChannelSelector(controller);
  setupProfileSelector(controller);
});

function setupProfileSelector(controller) {
  let profiles = {};
  let selected = '';
  renderSelector();

  // Retrieve profiles
  $.get('/profiles')
    .done((loadedProfiles) => {
      profiles = loadedProfiles;
      controller.profiles = profiles;
      selected = Object.keys(loadedProfiles)[0];
      renderSelector();
    })
    .fail((xhr) => {
      console.error('Profile retrieval failed:', xhr);
    });

  controller.on('updateTuning', ({profile}) => {
    selected = findKey(profiles, profile);
    renderSelector();
  });

  function updateSelection(value) {
    selected = value;
    renderSelector();
  }

  function renderSelector() {
    ReactDOM.render(
      <ProfileSelector
        profiles={profiles}
        selected={selected}
        onChange={updateSelection} />,
      document.getElementById('profile-selector')
    );
  }
}

function setupChannelSelector(controller) {
  let channels = [];
  let selected = '';
  renderSelector();

  // Retrieve channels
  $.get('/channels')
    .done((loadedChannels) => {
      channels = loadedChannels;
      controller.channels = channels;
      renderSelector();
    })
    .fail((xhr) => {
      console.error('Channel retrieval failed:', xhr);
    });

  controller.on('updateTuning', ({channel}) => {
    selected = channel;
    renderSelector();
  });

  function updateSelection(value) {
    selected = value;
    renderSelector();
  }

  function renderSelector() {
    ReactDOM.render(
      <ChannelSelector
        list={channels}
        selected={selected}
        onChange={updateSelection} />,
      document.getElementById('channel-selector')
    );
  }
}
