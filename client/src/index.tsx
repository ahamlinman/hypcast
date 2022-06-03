import React from "react";
import { createRoot } from "react-dom/client";

import App from "./App";
import { WebRTCProvider } from "./WebRTC";
import { TunerStatusProvider } from "./TunerStatus";

import "./index.scss";

const container = document.getElementById("root");
if (!container) {
  throw new Error('no element with ID "root"');
}

const root = createRoot(container);
root.render(
  <React.StrictMode>
    <WebRTCProvider>
      <TunerStatusProvider>
        <App />
      </TunerStatusProvider>
    </WebRTCProvider>
  </React.StrictMode>,
);
