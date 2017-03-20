import $ from 'jquery';
import { findKey } from 'lodash/object';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastClientController from './controller';
import ControllerBar from './ui/ControllerBar';

$(() => {
  let controller = new HypcastClientController();

  controller.on('transition', ({ fromState, toState }) => {
    console.debug(`state machine moving from ${fromState} to ${toState}`);
  });

  setupControllerBar(controller);
});

function setupControllerBar(controller) {
  let channels = [];
  let profiles = {};
  let selectedChannel = '';
  let selectedProfile = '';
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

  controller.on('updateTuning', ({ channel, profile }) => {
    selectedChannel = channel;
    selectedProfile = findKey(profiles, profile);
    render();
  });

  function handleChannelChanged(channelName) {
    selectedChannel = channelName;
    render();
  }

  function handleProfileChanged(profileName) {
    selectedProfile = profileName;
    render();
  }

  function render() {
    ReactDOM.render(
      <ControllerBar
        channels={channels}
        selectedChannel={selectedChannel}
        onChannelChanged={handleChannelChanged}
        profiles={profiles}
        selectedProfile={selectedProfile}
        onProfileChanged={handleProfileChanged} />,
      document.getElementById('controller-bar')
    );
  }
}
