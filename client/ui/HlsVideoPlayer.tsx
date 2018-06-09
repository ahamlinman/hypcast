import * as React from 'react';
import * as Hls from 'hls.js/dist/hls.light.js';

export default class HlsVideoPlayer extends React.Component<{ src: string }, {}> {
  private hls: Hls | null = null;
  private video: HTMLVideoElement | null = null;

  componentDidMount() {
    if (!this.video) {
      throw new Error('null <video> after HlsVideoPlayer mount');
    }

    this.hls = new Hls();
    this.hls.loadSource(this.props.src);
    this.hls.attachMedia(this.video);
    this.hls.on(Hls.Events.MANIFEST_PARSED, () => { if (this.video) { this.video.play(); } });
  }

  componentWillUnmount() {
    if (this.video) {
      this.video.pause();
      this.video = null;
    }

    if (this.hls) {
      this.hls.detachMedia();
      this.hls.destroy();
      this.hls = null;
    }
  }

  render() {
    return (
      <video ref={(video) => { this.video = video; }} controls={true} />
    );
  }
}
