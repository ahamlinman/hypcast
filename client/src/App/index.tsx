import React from "react";

import { useWebRTC } from "../WebRTC";
import rpc from "../rpc";

import "./index.scss";

import Header from "./Header";
import ChannelSelector from "./ChannelSelector";
import VideoPlayer from "./VideoPlayer";
import { useTunerStatus } from "../TunerStatus";

export default function App() {
  const webRTC = useWebRTC();
  const tunerStatus = useTunerStatus();

  return (
    <div className="AppContainer">
      <Header />
      <ChannelSelector
        selected={
          tunerStatus.Connection === "Connected" &&
          tunerStatus.State !== "Stopped"
            ? tunerStatus.ChannelName
            : undefined
        }
        onTune={(ChannelName) =>
          rpc("tune", { ChannelName }).catch(console.error)
        }
      />
      <VideoPlayer stream={webRTC.MediaStream} />
    </div>
  );
}
