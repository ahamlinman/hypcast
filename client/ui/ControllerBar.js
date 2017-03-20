import React from 'react';
import { findKey } from 'lodash/object';

import ChannelSelector from './ChannelSelector';
import ProfileSelector from './ProfileSelector';

export default class ControllerBar extends React.Component {
  render() {
    return (
      <form className="form-inline text-center lead" id="tuner">
	<div className="form-group">
	  <label htmlFor="channel">Watch </label>
	  <ChannelSelector
	    list={this.props.channels}
	    selected={this.props.tuneData.channel}
	    onChange={this.props.onChannelChanged} />
	</div>

	<div className="form-group">
	  <label htmlFor="profile"> at </label>
	  <ProfileSelector
	    profiles={this.props.profiles}
	    selected={findKey(this.props.profiles, this.props.tuneData.profile)}
	    onChange={this.props.onProfileChanged} />
	  <label htmlFor="profile"> quality</label>
	</div>
      </form>
    );
  }
}
