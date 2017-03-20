import React from 'react';
import { findKey } from 'lodash/object';

import ChannelSelector from './ChannelSelector';
import ProfileSelector from './ProfileSelector';

export default class ControllerBar extends React.Component {
  constructor(props) {
    super(props);

    this.handleSubmit = this.handleSubmit.bind(this);
    this.handleStop = this.handleStop.bind(this);
  }

  handleSubmit(event) {
    event.preventDefault();
    this.props.onTune();
  }

  handleStop(event) {
    event.preventDefault();
    this.props.onStop();
  }

  render() {
    return (
      <form className="form-inline text-center lead" id="tuner" onSubmit={this.handleSubmit}>
	<div className="form-group">
	  <label htmlFor="channel">Watch&nbsp;</label>
	  <ChannelSelector
	    list={this.props.channels}
	    selected={this.props.tuneData.channel}
	    onChange={this.props.onChannelChanged} />
	</div>
	&nbsp;
	<div className="form-group">
	  <label htmlFor="profile">at&nbsp;</label>
	  <ProfileSelector
	    profiles={this.props.profiles}
	    selected={findKey(this.props.profiles, this.props.tuneData.profile)}
	    onChange={this.props.onProfileChanged} />
	  <label htmlFor="profile">&nbsp;quality</label>
	</div>
	&nbsp;
	<button type="submit" className="btn btn-default">
	  <span className="glyphicon glyphicon-play"></span>
	</button>
	&nbsp;
	<button type="button" className="btn btn-default" onClick={this.handleStop}>
	  <span className="glyphicon glyphicon-stop"></span>
	</button>
      </form>
    );
  }
}
