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
    const options = Object.keys(this.props.profiles).map((name) => {
      const profile = this.props.profiles[name];
      return <option key={name} value={name}>{profile.description}</option>;
    });

    return (
      <select
          name='profile'
          className='form-control'
          disabled={!this.props.enabled}
          value={this.props.selected}
          onChange={this.handleChange}>
        {options}
      </select>
    );
  }
}
