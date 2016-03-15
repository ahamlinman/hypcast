import AzapError from './AzapError';
import { spawn } from 'child_process';
import { EventEmitter } from 'events';
import fs from 'fs';
import expandHomeDir from 'expand-home-dir';
import byline from 'byline';

class AzapTuner extends EventEmitter {
  constructor({
    channelsPath,
    adapter = 0,
    frontend = 0,
    demux = 0,
    device = '/dev/dvb/adapter0/dvr0'
  } = {}) {
    super();

    this.channelsPath = channelsPath;
    this.adapter = adapter;
    this.frontend = frontend;
    this.demux = demux;
    this.device = device;

    this._channel = null;
    this._locked = false;
  }

  tune(channel) {
    if (this._azap) {
      throw new Error('tuner is already tuned');
    }

    this._channel = channel;
    this._azap = this._spawnAzap(channel)
      .on('close', this._azapClose.bind(this));

    byline(this._azap.stdout).on('data', this._azapData.bind(this));

    this._stderrBuffer = new Buffer('');
    this._azap.stderr.on('data', (buf) => {
      this._stderrBuffer = Buffer.concat([this._stderrBuffer, buf]);
    });
  }

  stop() {
    if (this._azap) {
      this._azap.kill();
    } else {
      this.emit('stop');
    }
  }

  static loadChannels(path = '~/.azap/channels.conf') {
    return new Promise((resolve, reject) => {
      fs.readFile(expandHomeDir(path), 'ascii', (err, data) => {
        if (err) {
          reject(err);
        }

        let channelList = data.split('\n')
          .filter(entry => entry !== '')
          .map(entry => entry.split(':')[0])
          .filter((val, i, arr) => arr.indexOf(val) === i);

        resolve(channelList);
      });
    });
  }

  loadChannels(path = this.channelsPath) {
    return AzapTuner.loadChannels(path);
  }

  _spawnAzap(channel) {
    let azapOpts = [
      'azap', '-r',
      '-a', this.adapter,
      '-f', this.frontend,
      '-d', this.demux
    ];

    if (this.channelsPath) {
      azapOpts = azapOpts.concat('-c', this.channelsPath);
    }

    return spawn('stdbuf', ['-oL'].concat(azapOpts, channel));
  }

  _azapData(line) {
    if (!this._locked && line.toString().match(/FE_HAS_LOCK/)) {
      this.emit('lock', this._channel);
      this._locked = true;
    }
  }

  _azapClose(code) {
    delete this._azap;
    this._channel = null;
    this._locked = false;

    if (code) {
      this.emit('error', new AzapError(
            `azap failed with code ${code}`, this._stderrBuffer.toString()));
    } else {
      this.emit('stop');
    }
  }
}

export default AzapTuner;
