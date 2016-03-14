import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';
import Express from 'express';

let tuner = new AzapTuner();

tuner.on('error', (err) => console.log('azap error', err));
tuner.on('stop', () => console.log('azap stopped'));
tuner.on('lock', (channel) => {
  console.log('azap locked to', channel, '(press Ctrl-C to quit)');
});

let streamer = new HlsTunerStreamer(tuner);

streamer.on('transition', ({ fromState, toState }) => {
  console.log(`streamer moving from ${fromState} to ${toState}`);
});

streamer.on('error', (err) => console.log('streamer error', err));

streamer.tune({
  channel: process.argv[2],
  profile: {
    videoHeight: '480',
    videoBitrate: '768k',
    videoBufsize: '128k',
    videoPreset: 'fast',
    audioBitrate: '128k',
    audioProfile: 'aac_low',
  },
});

const app = Express();

app.use('/stream', (req, res, next) => {
  if (streamer.streamPath) {
    res.set('Access-Control-Allow-Origin', '*');
    Express.static(streamer.streamPath)(req, res, next);
  } else {
    next();
  }
});

let server = app.listen(9400, () => {
  console.log('Stream server started on :9400');
});

process.on('SIGINT', () => streamer.stop());

streamer.on('transition', ({ toState }) => {
  if (toState === 'inactive') {
    console.log('Closing stream server');
    server.close();
  }
});

server.on('close', () => console.log('Stream server closed'));
