import './hypcast.less';

import axios from 'axios';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastController from './controller';
import HypcastUi from './ui';

/**
 * This is the top level of the Hypcast client, which glues together its two
 * critical parts:
 *
 * - The controller, which both controls and synchronizes us with the server
 * - The UI tree containing all visual and interactive elements
 *
 * Because the UI is a React component, we need to be able to re-render it when
 * the state changes. That is handled quite easily by subscribing to the
 * server's state transition events. We also connect a couple of event handlers
 * from the React UI to controller methods.
 *
 * As a slightly special case, we also need to fire off requests for lists of
 * profiles and channels, since those come in via AJAX.
 */

document.addEventListener('DOMContentLoaded', () => {
  let controller = new HypcastController();

  controller.on('transition', render);

  controller.on('updateTuning', (update) => {
    tuneData = update;
    render();
  });

  let channels = [];
  let profiles = {};
  let tuneData = {
    channel: '',
    profile: {}
  };

  render();

  axios.get('/profiles')
    .then((response) => {
      profiles = response.data;

      if (tuneData.profile.description === undefined) {
        handleTuneDataChange({ profile: profiles[Object.keys(profiles)[0]] });
      } else {
        render();
      }
    })
    .catch((err) => {
      console.error('Profile retrieval failed:', err);
    });

  axios.get('/channels')
    .then((response) => {
      channels = response.data;

      if (tuneData.channel === '') {
        handleTuneDataChange({ channel: channels[0] });
      } else {
        render();
      }
    })
    .catch((err) => {
      console.error('Channel retrieval failed:', err);
    });

  function handleTuneDataChange(update) {
    tuneData = Object.assign({}, tuneData, update);
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
