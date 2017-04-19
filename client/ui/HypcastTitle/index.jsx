import React from 'react';
import styles from './styles.less';

export default class HypcastTitle extends React.Component {
  getClassNames() {
    return styles[this.props.state] || '';
  }

  render() {
    return <h1 className={this.getClassNames()}>hypcast</h1>;
  }
}
