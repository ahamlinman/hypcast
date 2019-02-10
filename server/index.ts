/* eslint-disable no-console */

import * as express from 'express';
import * as socketio from 'socket.io';
import * as path from 'path';
import * as fs from 'fs';
import { promisify } from 'util';

import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';

const channelsPath = path.resolve('config', 'channels.conf');
const tuner = new AzapTuner({ channelsPath });

tuner.on('error', (err) => console.log('[AzapTuner error]', err));

const streamer = new HlsTunerStreamer(tuner);
streamer.on('error', (err: Error) => console.log('[HlsTunerStreamer error]', err));
streamer.on('transition', ({ fromState, toState }: { fromState: string, toState: string }) => {
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

app.get('/profiles', async (_, res) => {
  const profilePath = path.resolve('config', 'profiles.json');
  const readFile = promisify(fs.readFile);

  try {
    const contents = await readFile(profilePath, 'utf-8');
    res.json(JSON.parse(contents));
  } catch (err) {
    res.status(500).send(err);
  }
});

app.get('/channels', async (_, res) => {
  try {
    const channels = await tuner.loadChannels();
    res.json(channels);
  } catch (err) {
    res.status(500).send(err);
  }
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

    const transitionHandler = ({ toState }: { toState: string }) => {
      socket.emit('transition', {
        toState,
        tuneData: streamer.tuneData,
      });
    };
    streamer.on('transition', transitionHandler);

    const errorHandler = (err: Error) => { socket.emit('hypcastError', err); };
    streamer.on('error', errorHandler);

    socket.on('disconnect', () => {
      console.log('client disconnected');
      streamer.off('transition', transitionHandler);
      streamer.off('error', errorHandler);
    });
  });
