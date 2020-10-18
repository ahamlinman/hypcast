import React, { FormEvent } from "react";

import { useWebRTC } from "../WebRTC";
import rpc from "../rpc";
import useConfig from "../useConfig";

import "./index.scss";

import Header from "./Header";

export default function App() {
  const webRTC = useWebRTC();

  return (
    <div className="AppContainer">
      <Header />
      <p>
        Controls:{" "}
        <ChannelSelector
          onTune={(ChannelName) =>
            rpc("tune", { ChannelName }).catch(console.error)
          }
        />
        <button onClick={() => rpc("stop").catch(console.error)}>Stop</button>
      </p>
      {webRTC.MediaStream ? <VideoPlayer stream={webRTC.MediaStream} /> : null}
    </div>
  );
}

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
