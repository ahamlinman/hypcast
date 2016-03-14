import AzapTuner from './AzapTuner';
import HlsTunerStreamer from './HlsTunerStreamer';

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

streamer.tune({ channel: 'SCI' });
setTimeout(() => streamer.tune({ channel: 'DISC' }), 500);
setTimeout(() => streamer.stop(), 6000);
