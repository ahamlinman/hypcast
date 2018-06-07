import * as React from 'react';
import * as Hls from 'hls.js/dist/hls.light.js';

export default class HlsVideoPlayer extends React.Component<{ src: string }, {}> {
  private hls: Hls;
  private video: HTMLVideoElement;

  componentDidMount() {
    this.hls = new Hls();
    this.hls.loadSource(this.props.src);
    this.hls.attachMedia(this.video);
    this.hls.on(Hls.Events.MANIFEST_PARSED, () => this.video.play());
  }

  componentWillUnmount() {
    this.video.pause();
    this.hls.detachMedia();
    this.hls.destroy();
  }

  render() {
    return (
      <video
        ref={(video) => { if (video !== null) { this.video = video; } }}
        controls={true}
      />
    );
  }
}
