import React from 'react';

export default App;

interface State {
  AudioDuration: number;
  VideoDuration: number;
}

interface EventData {
  Kind: string;
  Duration: number;
}

function App() {
  const [state, dispatch] = React.useReducer(reduceEvent, null, initState);

  React.useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.host}/hypcast/ws`)

    ws.onopen = () => {
      console.log('the websocket is open for business');
    };

    ws.onmessage = (evt) => {
      dispatch(JSON.parse(evt.data));
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <>
      <h1>It works!</h1>
      <p>Video Duration: {toSeconds(state.VideoDuration)}s</p>
      <p>Audio Duration: {toSeconds(state.AudioDuration)}s</p>
    </>
  );
}

function toSeconds(ns: number) {
  return (ns / 1_000_000_000).toFixed(3);
}

function initState(): State {
  return {
    AudioDuration: 0,
    VideoDuration: 0,
  };
}

function reduceEvent(state: State, eventData: EventData) {
  if (eventData.Kind === "video") {
    return {
      ...state,
      VideoDuration: state.VideoDuration + eventData.Duration,
    };
  }

  if (eventData.Kind === "audio") {
    return {
      ...state,
      AudioDuration: state.AudioDuration + eventData.Duration,
    };
  }

  throw new Error(`unrecognized event kind: ${eventData.Kind}`);
}
