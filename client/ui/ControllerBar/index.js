import React from 'react';
import { findKey } from 'lodash/object';

import ChannelSelector from './ChannelSelector';
import ProfileSelector from './ProfileSelector';

export default class ControllerBar extends React.Component {
  constructor(props) {
    super(props);

    this.handleChannelChanged = this.handleChannelChanged.bind(this);
    this.handleProfileChanged = this.handleProfileChanged.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
    this.handleStop = this.handleStop.bind(this);
  }

  handleChannelChanged(channel) {
    this.props.onTuneDataChange({ channel });
  }

  handleProfileChanged(profileName) {
    this.props.onTuneDataChange({ profile: this.props.profiles[profileName] });
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
	    onChange={this.handleChannelChanged} />
	</div>

	<div className="form-group">
	  <label htmlFor="profile">&nbsp;at&nbsp;</label>
	  <ProfileSelector
	    profiles={this.props.profiles}
	    selected={findKey(this.props.profiles, this.props.tuneData.profile)}
	    onChange={this.handleProfileChanged} />
	  <label htmlFor="profile">&nbsp;quality&nbsp;</label>
	</div>

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
