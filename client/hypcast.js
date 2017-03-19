import $ from 'jquery';
import { findKey } from 'lodash/object';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastClientController from './controller';
import ChannelSelector from './ui/ChannelSelector';

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
  setupChannelSelector();
});

function setupChannelSelector() {
  let channels = [];
  renderSelector();

  // Retrieve channels
  $.get('/channels')
    .done((loadedChannels) => {
      channels = loadedChannels;
      renderSelector();
    })
    .fail((xhr) => {
      console.error('Channel retrieval failed:', xhr);
    });

  function renderSelector() {
    ReactDOM.render(
      <ChannelSelector list={channels} />,
      document.getElementById('channel-selector')
    );
  }
}
