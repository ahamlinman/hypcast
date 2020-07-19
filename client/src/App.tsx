import React from "react";

import { useController } from "./Controller";

const App = () => {
  const [state] = useController();
  return (
    <>
      <h1>It works!</h1>
      <p>Status: {state.connected ? "Connected" : "Disconnected"}</p>
      {state.stream !== null ? <VideoPlayer stream={state.stream} /> : null}
    </>
  );
};

export default App;

const VideoPlayer = ({ stream }: { stream: MediaStream }) => {
  const videoElement = React.useRef<null | HTMLVideoElement>(null);

  React.useEffect(() => {
    if (videoElement.current === null) {
      return;
    }
    videoElement.current.srcObject = stream;
  }, [stream]);

  return (
    <video
      style={{ border: "1px solid black" }}
      ref={videoElement}
      autoPlay
      controls
    />
  );
};
