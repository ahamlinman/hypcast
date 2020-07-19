import React from "react";

import { useController } from "./Controller";

const App = () => {
  const [state] = useController();
  return (
    <>
      <h1>It works!</h1>
      <p>Status: {state.socketStatus}</p>
      <p>Video Duration: {formatDuration(state.videoDuration)}s</p>
      <p>Audio Duration: {formatDuration(state.audioDuration)}s</p>
    </>
  );
};

export default App;

const formatDuration = (ns: number) => (ns / 1e9).toFixed(3);
