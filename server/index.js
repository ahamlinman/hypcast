import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';
import express from 'express';
import socketio from 'socket.io';
import path from 'path';
import fs from 'fs';

let tuner = new AzapTuner({
  channelsPath: path.resolve('config', 'channels.conf')
});

tuner.on('error', err => console.log('[AzapTuner error]', err));

let streamer = new HlsTunerStreamer(tuner);
streamer.on('error', err => console.log('[HlsTunerStreamer error]', err));
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
  let profilePath = path.resolve('config', 'profiles.json');
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
      res.status(500).send(err)
    });
});

let server = app.listen(9400, () => {
  console.log('hypcast server started on *:9400');
});

let io = socketio(server)
  .on('connection', (socket) => {
    console.log('client connected');
    socket.emit('transition', {
      toState: streamer.state,
      tuneData: streamer.tuneData,
    });

    socket.on('tune', options => streamer.tune(options));
    socket.on('stop', () => streamer.stop());

    let transSub = streamer.on('transition', ({ toState }) => {
      socket.emit('transition', {
        toState,
        tuneData: streamer.tuneData,
      });
    });

    let errSub = streamer.on('error', (err) => {
      socket.emit('hypcastError', err);
    });

    socket.on('disconnect', () => {
      console.log('client disconnected');
      transSub.off();
      errSub.off();
    });
  });
