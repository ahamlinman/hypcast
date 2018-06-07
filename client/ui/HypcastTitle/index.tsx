import * as React from 'react';
import * as styles from './styles.less';

interface HypcastTitleProps {
  state: string;
}

export default class HypcastTitle extends React.Component<HypcastTitleProps, {}> {
  getClassNames() {
    return styles[this.props.state] || '';
  }

  render() {
    return <h1 className={this.getClassNames()}>hypcast</h1>;
  }
}
