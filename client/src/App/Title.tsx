import React from "react";
import { Helmet } from "react-helmet";

import { useTunerStatus } from "../TunerStatus";

export default function Title() {
  const tunerStatus = useTunerStatus();

  const titleText =
    tunerStatus.Connection === "Connected" && tunerStatus.State !== "Stopped"
      ? `${tunerStatus.ChannelName} | Hypcast`
      : "Hypcast";

  return (
    <Helmet>
      <title>{titleText}</title>
    </Helmet>
  );
}
