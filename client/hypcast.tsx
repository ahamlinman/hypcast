/* eslint-disable no-console */

import './hypcast.less';

import * as React from 'react';
import * as ReactDOM from 'react-dom';

import HypcastController from './controller';
import HypcastUi, { ProfileSet, TuneData, TuneDataChange } from './ui';

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
  const controller = new HypcastController();

  controller.on('transition', render);

  controller.on('updateTuning', (update: TuneData) => {
    tuneData = update;
    render();
  });

  let channels: string[] = [];
  let profiles: ProfileSet = {};
  let tuneData: TuneData = {
    channel: '',
    profile: { description: '' },
  };

  render();

  (async function getProfiles() {
    const response = await fetch('/profiles');

    if (!response.ok) {
      console.error('Profile retrieval failed', response);
      throw new Error('Profile retrieval failed');
    }

    profiles = await response.json();

    if (tuneData.profile.description === undefined) {
      handleTuneDataChange({ profile: profiles[Object.keys(profiles)[0]] });
    } else {
      render();
    }
  })();

  (async function getChannels() {
    const response = await fetch('/channels');

    if (!response.ok) {
      console.error('Channel retrieval failed', response);
      throw new Error('Channel retrieval failed');
    }

    channels = await response.json();

    if (tuneData.channel === '') {
      handleTuneDataChange({ channel: channels[0] });
    } else {
      render();
    }
  })();

  function handleTuneDataChange(update: TuneDataChange) {
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
