import { spawn, ChildProcess } from 'child_process';
import { EventEmitter } from 'events';
import { promisify } from 'util';
import * as fs from 'fs';
import * as byline from 'byline';

import AzapError from './AzapError';

export default class AzapTuner extends EventEmitter {
  public readonly channelsPath: string;
  public readonly adapter: number;
  public readonly frontend: number;
  public readonly demux: number;
  public readonly device: string;

  private _channel: string | null = null;
  private _locked: boolean = false;
  private _azap: ChildProcess | null = null;
  private _stderrBuf: Buffer | null = null;

  constructor({
    channelsPath = 'channels.conf',
    adapter = 0,
    frontend = 0,
    demux = 0,
    device = '/dev/dvb/adapter0/dvr0',
  } = {}) {
    super();

    this.channelsPath = channelsPath;
    this.adapter = adapter;
    this.frontend = frontend;
    this.demux = demux;
    this.device = device;
  }

  tune(channel) {
    if (this._azap) {
      throw new Error('tuner is already tuned');
    }

    this._channel = channel;
    this._azap = this._spawnAzap(channel)
      .on('close', this._azapClose.bind(this));

    byline(this._azap.stdout).on('data', this._azapData.bind(this));

    this._stderrBuf = null;
    this._azap.stderr.on('data', (buf: Buffer) => {
      this._stderrBuf = this._stderrBuf ? Buffer.concat([this._stderrBuf, buf]) : buf;
    });
  }

  stop() {
    if (this._azap) {
      this._azap.kill();
    } else {
      this.emit('stop');
    }
  }

  static async loadChannels(path) {
    const readFile = promisify(fs.readFile);
    const data = await readFile(path, 'ascii');

    return data.split('\n')
      .filter((entry) => entry !== '')
      .map((entry) => entry.split(':')[0])
      .filter((val, i, arr) => arr.indexOf(val) === i);
  }

  loadChannels(path = this.channelsPath) {
    return AzapTuner.loadChannels(path);
  }

  _spawnAzap(channel) {
    let azapOpts = [
      'azap', '-r',
      '-a', this.adapter.toString(),
      '-f', this.frontend.toString(),
      '-d', this.demux.toString(),
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
        `azap failed with code ${code}`,
        this._stderrBuf ? this._stderrBuf.toString() : ''));
    } else {
      this.emit('stop');
    }
  }
}
