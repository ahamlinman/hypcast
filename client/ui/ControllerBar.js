import React from 'react';

import ChannelSelector from './ChannelSelector';
import ProfileSelector from './ProfileSelector';

export default class ControllerBar extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      selectedChannel: '',
      selectedProfile: ''
    };

    this.handleChannelChanged = this.handleChannelChanged.bind(this);
    this.handleProfileChanged = this.handleProfileChanged.bind(this);
  }

  handleChannelChanged(selectedChannel) {
    this.setState({selectedChannel});
  }

  handleProfileChanged(selectedProfile) {
    this.setState({selectedProfile});
  }

  render() {
    return (
      <form className="form-inline text-center lead" id="tuner">
	<div className="form-group">
	  <label htmlFor="channel">Watch </label>
	  <ChannelSelector
	    list={this.props.channels}
	    selected={this.state.selectedChannel}
	    onChange={this.handleChannelChanged} />
	</div>

	<div className="form-group">
	  <label htmlFor="profile"> at </label>
	  <ProfileSelector
	    profiles={this.props.profiles}
	    selected={this.state.selectedProfile}
	    onChange={this.handleProfileChanged} />
	  <label htmlFor="profile"> quality</label>
	</div>
      </form>
    );
  }
}
