import './hypcast.less';

import axios from 'axios';
import React from 'react';
import ReactDOM from 'react-dom';

import HypcastController from './controller';
import HypcastUi from './ui';

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
      render();
    })
    .catch((err) => {
      console.error('Profile retrieval failed:', err);
    });

  axios.get('/channels')
    .then((response) => {
      channels = response.data;
      render();
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
