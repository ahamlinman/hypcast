import React from 'react';

export default class ProfileSelector extends React.Component {
  render() {
    let options = Object.keys(this.props.profiles).map((name) => {
      let profile = this.props.profiles[name];
      return <option value={name}>{profile.description}</option>;
    });

    return (
      <select
	  name="profile"
	  id="profile"
	  className="form-control">
	{options}
      </select>
    );
  }
}
