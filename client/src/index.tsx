import React from "react";
import ReactDOM from "react-dom";

import App from "./App";
import { WebRTCProvider } from "./WebRTC";
import { TunerStatusProvider } from "./TunerStatus";

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
