class AzapTuner {
  constructor({
    channelsPath,
    adapter = 0,
    frontend = 0,
    demux = 0,
    device = '/dev/dvb/adapter0/dvr0'
  } = {}) {
    this.channelsPath = channelsPath;
    this.adapter = adapter;
    this.frontend = frontend;
    this.demux = demux;
    this.device = device;
  }
}

export default AzapTuner;
