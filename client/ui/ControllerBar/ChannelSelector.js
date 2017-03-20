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
    let options = this.props.list.map((channel) => {
      return <option value={channel}>{channel}</option>;
    });

    return (
      <select
	  name="channel"
	  id="channel"
	  value={this.props.selected}
	  className="form-control"
	  onChange={this.handleChange}>
	{options}
      </select>
    );
  }
}
