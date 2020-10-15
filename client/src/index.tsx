import React from "react";
import ReactDOM from "react-dom";

import App from "./App";
import { Controller } from "./Controller";
import { TunerStatusProvider } from "./TunerStatus";

ReactDOM.render(
  <React.StrictMode>
    <Controller>
      <TunerStatusProvider>
        <App />
      </TunerStatusProvider>
    </Controller>
  </React.StrictMode>,
  document.getElementById("root"),
);
