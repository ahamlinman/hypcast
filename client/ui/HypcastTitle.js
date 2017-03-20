import React from 'react';

export default class HypcastTitle extends React.Component {
  getClassName() {
    let stateMap = {
      connecting: 'text-muted',
      tuning: 'hyp-tuning',
      buffering: 'hyp-buffering',
      active: 'text-success'
    };

    return stateMap[this.props.state] || '';
  }

  render() {
    return <h1 className={this.getClassName()}>hypcast</h1>;
  }
}
