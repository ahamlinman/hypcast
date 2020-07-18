import React from "react";

import { Context as ControllerContext } from "./Controller";

const App = () => {
  const controller = React.useContext(ControllerContext);
  return (
    <>
      <h1>It works!</h1>
      <p>Status: {controller.status}</p>
      <p>Video Duration: {formatDuration(controller.videoDuration)}s</p>
      <p>Audio Duration: {formatDuration(controller.audioDuration)}s</p>
    </>
  );
};

export default App;

const formatDuration = (ns: number) => (ns / 1e9).toFixed(3);
