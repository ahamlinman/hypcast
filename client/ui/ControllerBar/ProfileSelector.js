import React from 'react';

export default class ProfileSelector extends React.Component {
  constructor(props) {
    super(props);
    this.handleChange = this.handleChange.bind(this);
  }

  handleChange(event) {
    this.props.onChange(event.target.value);
  }

  render() {
    let options = Object.keys(this.props.profiles).map((name) => {
      let profile = this.props.profiles[name];
      return <option value={name}>{profile.description}</option>;
    });

    return (
      <select
	  name="profile"
	  id="profile"
	  value={this.props.selected}
	  className="form-control"
	  onChange={this.handleChange}>
	{options}
      </select>
    );
  }
}
