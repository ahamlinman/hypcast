import React, { FormEvent } from "react";

import { useController } from "./Controller";

const App = () => {
  const controller = useController();

  return (
    <>
      <h1>It works!</h1>
      <p>Connection Status: {controller.connectionState.status}</p>
      <p>Tuner Status: {controller.tunerState?.status}</p>
      <p>Current Channel: {controller.requestedChannelName || "(unknown)"}</p>
      <p>
        Change Channel:{" "}
        <ChannelSelector
          onTune={async (name: string) => {
            controller.changeChannel(name);
          }}
        />
      </p>
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
  onTune,
}: {
  onTune: (ch: string) => Promise<void>;
}) => {
  const channelNames = useChannelNames();
  const [selected, setSelected] = React.useState<undefined | string>();
  const [forceDisabled, setForceDisabled] = React.useState(false);

  React.useEffect(() => {
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
    <form style={{ display: "inline" }} onSubmit={handleTune}>
      <select
        name="channel"
        disabled={disabled}
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
      <button type="submit" disabled={disabled}>
        Tune
      </button>
    </form>
  );
};

const useChannelNames = (): undefined | string[] | Error => {
  const [result, setResult] = React.useState<undefined | string[] | Error>();

  React.useEffect(() => {
    const startFetch = async () => {
      try {
        const result = await fetch("/config/channels");
        const channels: string[] = await result.json();
        setResult(channels);
      } catch (e) {
        setResult(e);
      }
    };

    startFetch();
  }, []);

  return result;
};
