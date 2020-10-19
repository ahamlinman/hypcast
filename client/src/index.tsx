import React from "react";
import ReactDOM from "react-dom";

import App from "./App";
import { WebRTCProvider } from "./WebRTC";
import { TunerStatusProvider } from "./TunerStatus";

import "./index.scss";

ReactDOM.render(
  <React.StrictMode>
    <WebRTCProvider>
      <TunerStatusProvider>
        <App />
      </TunerStatusProvider>
    </WebRTCProvider>
  </React.StrictMode>,
  document.getElementById("root"),
);
