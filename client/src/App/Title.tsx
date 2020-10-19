import React from "react";
import { Helmet } from "react-helmet";

import { useTunerStatus } from "../TunerStatus";

export default function Title() {
  const tunerStatus = useTunerStatus();

  return (
    <Helmet>
      <title>
        {tunerStatus.Connection === "Connected" &&
        tunerStatus.State !== "Stopped"
          ? `${tunerStatus.ChannelName} | Hypcast`
          : "Hypcast"}
      </title>
    </Helmet>
  );
}
