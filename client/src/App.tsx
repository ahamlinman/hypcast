import React, { FormEvent } from "react";

import { useController } from "./Controller";

const App = () => {
  const controller = useController();

  return (
    <>
      <h1>It works!</h1>
      <p>Connection Status: {controller.connectionState.status}</p>
      <p>Tuner Status: {controller.tunerState?.status}</p>
      <p>Now Watching: {controller.currentChannelName || "(none)"}</p>
      <p>
        Controls: <ChannelSelector onTune={changeChannel} />
        <button onClick={turnOff}>Stop</button>
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

const useChannelNames = (): undefined | string[] | Error => {
  const [result, setResult] = React.useState<undefined | string[] | Error>();

  React.useEffect(() => {
    const startFetch = async () => {
      try {
        const result = await fetch("/api/config/channels");
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

function changeChannel(name: string) {
  return callRPC("tune", { ChannelName: name }).catch(console.error);
}

function turnOff() {
  return callRPC("stop").catch(console.error);
}

async function callRPC(name: string, params?: { [k: string]: any }) {
  const response = await fetch(`/api/rpc/${name}`, {
    method: "POST",
    body: params ? JSON.stringify(params) : undefined,
    headers: params
      ? new Headers({ "Content-Type": "application/json" })
      : undefined,
  });

  if (!response.ok) {
    const body = await response.json();
    throw new Error(body.Error);
  }
}
