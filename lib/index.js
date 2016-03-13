import AzapTuner from './AzapTuner';

let tuner = new AzapTuner();

tuner.on('error', (err) => console.log('azap error', err));
tuner.on('stop', () => console.log('azap stopped'));
tuner.on('lock', (channel) => {
  console.log('azap locked to', channel, '(press Ctrl-C to quit)');
});

tuner.tune(process.argv[2]);
