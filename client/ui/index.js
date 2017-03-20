import React from 'react';

import HypcastTitle from './HypcastTitle';
import HlsVideoPlayer from './HlsVideoPlayer';
import ControllerBar from './ControllerBar';

export default class HypcastUi extends React.Component {
  getVideoElement() {
    return (
      this.props.state === 'active' ?
      <HlsVideoPlayer src="/stream/stream.m3u8" /> :
      <span />
    );

  }

  getControllerBarEnabled() {
    return this.props.state !== 'connecting';
  }

  render() {
    return (
      <div>
	<div className="page-header text-center">
	  <HypcastTitle state={this.props.state} />
	</div>

	<div className="row">
	  <div className="col-xs-12 col-sm-10 col-sm-push-1 col-md-8 col-md-push-2">
	    {this.getVideoElement()}
	  </div>
	</div>

	<div className="row">
	  <div className="col-xs-10 col-xs-push-1">
	    <ControllerBar
	      enabled={this.getControllerBarEnabled()}
	      channels={this.props.channels}
	      profiles={this.props.profiles}
	      tuneData={this.props.tuneData}
	      onTuneDataChange={this.props.onTuneDataChange}
	      onTune={this.props.onTune}
	      onStop={this.props.onStop} />
	  </div>
	</div>
      </div>
    );
  }
}
