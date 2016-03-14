import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';
import express from 'express';
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

let server = app.listen(9400, () => {
  console.log('hypcast server started on *:9400');
});
