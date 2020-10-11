import React from "react";

import { useController } from "./Controller";

const App = () => {
  const controller = useController();

  return (
    <>
      <h1>It works!</h1>
      <p>Connection Status: {controller.connectionState.status}</p>
      <p>Tuner Status: {controller.tunerState?.status}</p>
      {controller.requestedChannelName ? (
        <p>Current Channel: {controller.requestedChannelName}</p>
      ) : null}
      {controller.channelList ? (
        <>
          Change Channel:{" "}
          <ChannelSelector
            channelNames={controller.channelList}
            onTune={async (name: string) => {
              controller.changeChannel(name);
            }}
          />
          <br />
        </>
      ) : null}
      {controller.mediaStream ? (
        <VideoPlayer stream={controller.mediaStream} />
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

const ChannelSelector = ({
  channelNames,
  onTune,
}: {
  channelNames: string[];
  onTune: (ch: string) => Promise<void>;
}) => {
  const [selected, setSelected] = React.useState(channelNames[0]);
  const [disabled, setDisabled] = React.useState(false);

  const handleTune = async () => {
    setDisabled(true);
    try {
      await onTune(selected);
    } catch (e) {
      console.error("Tune request failed", e);
    } finally {
      setDisabled(false);
    }
  };

  return (
    <>
      <select
        disabled={disabled}
        value={selected}
        onChange={(evt) => setSelected(evt.currentTarget.value)}
      >
        {channelNames.map((ch) => (
          <option key={ch} value={ch}>
            {ch}
          </option>
        ))}
      </select>
      <button disabled={disabled} onClick={handleTune}>
        Tune
      </button>
    </>
  );
};
