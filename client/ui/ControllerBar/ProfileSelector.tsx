import * as React from 'react';

export interface Profile {
  description: string;
}

interface ProfileSelectorProps {
  enabled: boolean;
  onChange: (value: string) => void;
  selected: string | undefined;
  profiles: { [name: string]: Profile };
}

export default class ProfileSelector extends React.Component<ProfileSelectorProps, {}> {
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
