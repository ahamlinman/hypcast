import * as React from 'react';

interface ChannelSelectorProps {
  enabled: boolean;
  onChange: (value: string) => void;
  selected: string | undefined;
  channels: string[];
}

export default class ChannelSelector extends React.Component<ChannelSelectorProps, {}> {
  constructor(props: ChannelSelectorProps) {
    super(props);
    this.handleChange = this.handleChange.bind(this);
  }

  handleChange(event: React.ChangeEvent<HTMLSelectElement>) {
    this.props.onChange(event.target.value);
  }

  render() {
    const options = this.props.channels
      .map((channel) => <option key={channel} value={channel}>{channel}</option>);

    return (
      <select
          name='channel'
          className='form-control'
          disabled={!this.props.enabled}
          value={this.props.selected}
          onChange={this.handleChange}>
        {options}
      </select>
    );
  }
}
