import React from 'react';

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
	    selected={this.props.selectedChannel}
	    onChange={this.props.onChannelChanged} />
	</div>

	<div className="form-group">
	  <label htmlFor="profile"> at </label>
	  <ProfileSelector
	    profiles={this.props.profiles}
	    selected={this.props.selectedProfile}
	    onChange={this.props.onProfileChanged} />
	  <label htmlFor="profile"> quality</label>
	</div>
      </form>
    );
  }
}
