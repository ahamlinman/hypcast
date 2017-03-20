import React from 'react';
import styles from './styles.css';

export default class HypcastTitle extends React.Component {
  getClassNames() {
    return styles[this.props.state] || '';
  }

  render() {
    return <h1 className={this.getClassNames()}>hypcast</h1>;
  }
}
