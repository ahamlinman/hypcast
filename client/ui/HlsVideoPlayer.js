import React from 'react';
import Hls from 'hls.js';

export default class HlsVideoPlayer extends React.Component {
  componentDidMount() {
    this.hls = new Hls();
    this.hls.loadSource(this.props.src);
    this.hls.attachMedia(this.video);
    this.hls.on(Hls.Events.MANIFEST_PARSED, () => this.video.play());
  }

  componentWillUnmount() {
    this.video.pause();
    this.hls.detachMedia(this.video);
    this.hls.destroy();
  }

  render() {
    return <video ref={(video) => { this.video = video; }} controls={true} />;
  }
}
