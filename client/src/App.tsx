import React, { FormEvent } from "react";

import { useController } from "./Controller";
import { useTunerStatus, Status as TunerStatus } from "./TunerStatus";

import rpc from "./rpc";
import useConfig from "./useConfig";

const App = () => {
  const controller = useController();
  const tunerStatus = useTunerStatus();

  return (
    <>
      <h1>It works!</h1>
      <p>WebRTC Status: {controller.connectionState.status}</p>
      <TunerStatusDisplay status={tunerStatus} />
      <p>
        Controls:{" "}
        <ChannelSelector
          onTune={(ChannelName) =>
            rpc("tune", { ChannelName }).catch(console.error)
          }
        />
        <button onClick={() => rpc("stop").catch(console.error)}>Stop</button>
      </p>
      {controller.mediaStream ? (
        <VideoPlayer stream={controller.mediaStream} />
      ) : null}
    </>
  );
};

export default App;

const TunerStatusDisplay = ({ status }: { status: TunerStatus }) => (
  <p>Tuner Status: {tunerStatusToString(status)}</p>
);

const tunerStatusToString = (status: TunerStatus) => {
  if (status.Connection !== "Connected") {
    return `(${status.Connection})`;
  }

  if (status.State === "Stopped") {
    if (status.Error !== undefined) {
      return `${status.State} (${status.Error})`;
    }
    return status.State;
  }

  return `${status.State} ${status.ChannelName}`;
};

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
  onTune,
}: {
  onTune: (ch: string) => Promise<void>;
}) => {
  const channelNames = useConfig<string[]>("channels");

  const [selected, setSelected] = React.useState<undefined | string>();
  const [forceDisabled, setForceDisabled] = React.useState(false);

  React.useEffect(() => {
    if (channelNames instanceof Error) {
      console.error(channelNames);
    }
    if (channelNames instanceof Array) {
      setSelected((s) => (s === undefined ? channelNames[0] : s));
    }
  }, [channelNames]);

  const disabled =
    forceDisabled ||
    channelNames === undefined ||
    channelNames instanceof Error;

  const handleTune = async (evt: FormEvent) => {
    evt.preventDefault();

    if (selected === undefined) {
      throw new Error("tried to tune before channels loaded");
    }

    setForceDisabled(true);
    try {
      await onTune(selected);
    } catch (e) {
      console.error("Tune request failed", e);
    } finally {
      setForceDisabled(false);
    }
  };

  return (
    <>
      <select
        name="channel"
        value={selected}
        onChange={(evt) => setSelected(evt.currentTarget.value)}
      >
        {channelNames instanceof Array
          ? channelNames.map((ch) => (
              <option key={ch} value={ch}>
                {ch}
              </option>
            ))
          : null}
      </select>
      <button disabled={disabled} onClick={handleTune}>
        Tune
      </button>
    </>
  );
};
