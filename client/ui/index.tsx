import * as React from 'react';
import * as CSSTransition from 'react-transition-group/CSSTransition';

import HypcastTitle from './HypcastTitle';
import HlsVideoPlayer from './HlsVideoPlayer';
import ControllerBar, { ProfileSet, TuneData, TuneDataChange, TuneDataProps } from './ControllerBar';

import * as videoTransitions from './videoTransitions.less';

export { ProfileSet, TuneData, TuneDataChange };

interface HypcastUiProps extends TuneDataProps {
  state: string;
  channels: string[];
  profiles: ProfileSet;
}

export default class HypcastUi extends React.Component<HypcastUiProps, {}> {
  getVideoElement() {
    const active = this.props.state === 'active';

    return (
      <CSSTransition in={active} mountOnEnter={true} unmountOnExit={true}
          classNames={videoTransitions} timeout={350}>
        <HlsVideoPlayer src='/stream/stream.m3u8' />
      </CSSTransition>
    );
  }

  getControllerBarEnabled() {
    return this.props.state !== 'connecting';
  }

  render() {
    return (
      <div>
        <div className='page-header text-center'>
          <HypcastTitle state={this.props.state} />
        </div>

        <div className='row'>
          <div className='col-xs-12 col-sm-10 col-sm-push-1 col-md-8 col-md-push-2'>
            {this.getVideoElement()}
          </div>
        </div>

        <div className='row'>
          <div className='col-xs-10 col-xs-push-1'>
            <ControllerBar
              enabled={this.getControllerBarEnabled()}
              channels={this.props.channels}
              profiles={this.props.profiles}
              tuneData={this.props.tuneData}
              onTuneDataChange={this.props.onTuneDataChange}
              onTune={this.props.onTune}
              onStop={this.props.onStop} />
          </div>
        </div>
      </div>
    );
  }
}
