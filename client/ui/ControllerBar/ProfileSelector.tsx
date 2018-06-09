import * as React from 'react';

export interface Profile {
  description: string;
}

export type ProfileSet = { [name: string]: Profile };

interface ProfileSelectorProps {
  enabled: boolean;
  onChange: (value: string) => void;
  selected: string | undefined;
  profiles: ProfileSet;
}

export default class ProfileSelector extends React.Component<ProfileSelectorProps, {}> {
  constructor(props: ProfileSelectorProps) {
    super(props);
    this.handleChange = this.handleChange.bind(this);
  }

  handleChange(event: React.ChangeEvent<HTMLSelectElement>) {
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
