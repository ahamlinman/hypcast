/* eslint-disable no-console */

import express from 'express';
import socketio from 'socket.io';
import path from 'path';
import fs from 'fs';

import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';

const channelsPath = path.resolve('config', 'channels.conf');
const tuner = new AzapTuner({ channelsPath });

tuner.on('error', (err) => console.log('[AzapTuner error]', err));

const streamer = new HlsTunerStreamer(tuner);
streamer.on('error', (err) => console.log('[HlsTunerStreamer error]', err));
streamer.on('transition', ({ fromState, toState }) => {
  console.log(`streamer moving from ${fromState} to ${toState}`);
});

const app = express();

app.use('/', express.static(path.join(__dirname, '../client')));

app.use('/stream', (req, res, next) => {
  if (streamer.streamPath) {
    res.set('Access-Control-Allow-Origin', '*');
    express.static(streamer.streamPath)(req, res, next);
  } else {
    next();
  }
});

app.get('/profiles', (req, res) => {
  const profilePath = path.resolve('config', 'profiles.json');
  fs.readFile(profilePath, (err, contents) => {
    if (err) {
      res.status(500).send(err);
    } else {
      res.json(JSON.parse(contents));
    }
  });
});

app.get('/channels', (req, res) => {
  tuner.loadChannels()
    .then((channels) => {
      res.json(channels);
    })
    .catch((err) => {
      res.status(500).send(err);
    });
});

const server = app.listen(9400, () => {
  console.log('hypcast server started on *:9400');
});

socketio(server)
  .on('connection', (socket) => {
    console.log('client connected');
    socket.emit('transition', {
      toState: streamer.state,
      tuneData: streamer.tuneData,
    });

    socket.on('tune', (options) => streamer.tune(options));
    socket.on('stop', () => streamer.stop());

    const transSub = streamer.on('transition', ({ toState }) => {
      socket.emit('transition', {
        toState,
        tuneData: streamer.tuneData,
      });
    });

    const errSub = streamer.on('error', (err) => {
      socket.emit('hypcastError', err);
    });

    socket.on('disconnect', () => {
      console.log('client disconnected');
      transSub.off();
      errSub.off();
    });
  });
