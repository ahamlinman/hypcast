import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';
import express from 'express';
import socketio from 'socket.io';
import path from 'path';
import fs from 'fs';

let tuner = new AzapTuner();
let streamer = new HlsTunerStreamer(tuner);

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
  let profilePath = path.join(__dirname, '../client/profiles.json');
  fs.readFile(profilePath, (err, contents) => {
    if (err) {
      res.status(500).send(err);
    } else {
      res.json(JSON.parse(contents));
    }
  });
});

app.get('/channels', (req, res) => {
  AzapTuner.loadChannels()
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
    socket.emit('transition', { toState: streamer.state });

    socket.on('tune', options => streamer.tune(options));

    let transSub = streamer.on('transition', (options) => {
      socket.emit('transition', options);
    });

    socket.on('disconnect', () => {
      console.log('client disconnected');
      transSub.off();
    });
  });
