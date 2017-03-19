import React from 'react';

export default class ChannelSelector extends React.Component {
  render() {
    let options = this.props.list.map((channel) => {
      return <option value={channel}>{channel}</option>;
    });

    return (
      <select
	  name="channel"
	  id="channel"
	  className="form-control">
	{options}
      </select>
    );
  }
}
