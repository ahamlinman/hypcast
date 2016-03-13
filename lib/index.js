import AzapTuner from './AzapTuner';

let tuner = new AzapTuner();

tuner.on('error', (err) => console.log('azap error', err));
tuner.on('stop', () => console.log('azap stopped'));

tuner.tune(process.argv[2]);
tuner.on('lock', (channel) => {
  console.log('azap locked to', channel, '(killing in 3 seconds)');
  setTimeout(() => tuner.stop(), 3000);
});
