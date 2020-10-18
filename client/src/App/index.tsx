import React from "react";

import { useWebRTC } from "../WebRTC";
import rpc from "../rpc";

import "./index.scss";

import Header from "./Header";
import ChannelSelector from "./ChannelSelector";
import VideoPlayer from "./VideoPlayer";

export default function App() {
  const webRTC = useWebRTC();

  return (
    <div className="AppContainer">
      <Header />
      <ChannelSelector
        onTune={(ChannelName) =>
          rpc("tune", { ChannelName }).catch(console.error)
        }
      />
      <VideoPlayer stream={webRTC.MediaStream} />
    </div>
  );
}
