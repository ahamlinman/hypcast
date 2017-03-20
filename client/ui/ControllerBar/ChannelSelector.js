import React from 'react';

export default class ChannelSelector extends React.Component {
  constructor(props) {
    super(props);
    this.handleChange = this.handleChange.bind(this);
  }

  handleChange(event) {
    this.props.onChange(event.target.value);
  }

  render() {
    let options = this.props.channels.map((channel) => {
      return <option value={channel}>{channel}</option>;
    });

    return (
      <select
	  name="channel"
	  className="form-control"
	  disabled={!this.props.enabled}
	  value={this.props.selected}
	  onChange={this.handleChange}>
	{options}
      </select>
    );
  }
}
