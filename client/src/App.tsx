import React from 'react';

export default App;

function App() {
  const wsRef = React.useRef<WebSocket>();
  React.useEffect(() => {
    wsRef.current = new WebSocket(`ws://${window.location.host}/hypcast/ws`)

    wsRef.current.onopen = () => {
      console.log('the websocket is open for business');
    };

    wsRef.current.onmessage = (evt) => {
      console.log('got a blob', evt.data);
    };

    return () => {};
  });

  return <h1>It works!</h1>;
}
