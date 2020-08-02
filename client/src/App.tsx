import React from "react";

import { useController } from "./Controller";

const App = () => {
  const controller = useController();

  return (
    <>
      <h1>It works!</h1>
      <p>Connection Status: {controller.connectionState.status}</p>
      <p>Tuner Status: {controller.tunerState?.status}</p>
      {controller.mediaStream ? (
        <VideoPlayer stream={controller.mediaStream} />
      ) : null}
      <br />
      {controller.channelList && controller.requestedChannelName ? (
        <Selector
          options={controller.channelList}
          value={controller.requestedChannelName}
          onChange={controller.changeChannel}
        />
      ) : null}
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

const Selector = ({
  options,
  value,
  onChange,
}: {
  options: string[];
  value: string;
  onChange: (v: string) => void;
}) => (
  <select value={value} onChange={(evt) => onChange(evt.currentTarget.value)}>
    {options.map((opt) => (
      <option key={opt} value={opt}>
        {opt}
      </option>
    ))}
  </select>
);
