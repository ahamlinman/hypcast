import React from "react";
import ReactDOM from "react-dom";

import App from "./App";
import Controller from "./Controller";

ReactDOM.render(
  <React.StrictMode>
    <Controller>
      <App />
    </Controller>
  </React.StrictMode>,
  document.getElementById("root"),
);
