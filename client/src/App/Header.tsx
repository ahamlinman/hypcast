import React from "react";

import { useWebRTC } from "../WebRTC";
import { useTunerStatus, Status as TunerStatus } from "../TunerStatus";

export default function Header() {
  const webRTC = useWebRTC();
  const tunerStatus = useTunerStatus();

  return (
    <header className="AppHeader">
      <h1>hypcast</h1>
      <StatusIndicator
        description={`WebRTC ${webRTC.Connection.Status}`}
        active={webRTC.Connection.Status === "Connected"}
      />
      <StatusIndicator
        description={`Tuner ${tunerStatusToString(tunerStatus)}`}
        active={
          tunerStatus.Connection === "Connected" &&
          tunerStatus.State === "Playing"
        }
      />
    </header>
  );
}

function StatusIndicator({
  active,
  description,
}: {
  active: boolean;
  description: string;
}) {
  return (
    <div className="AppHeaderStatusIndicator">
      <div
        className={`AppHeaderStatusIndicatorDot ${active ? "Active" : ""}`}
      ></div>
      <span className="AppHeaderStatusIndicatorDescription">{description}</span>
    </div>
  );
}

const tunerStatusToString = (status: TunerStatus) => {
  if (status.Connection !== "Connected") {
    return `${status.Connection}`;
  }

  if (status.State === "Stopped") {
    if (status.Error !== undefined) {
      return `${status.State} (${status.Error})`;
    }
    return status.State;
  }

  return `${status.State} ${status.ChannelName}`;
};
