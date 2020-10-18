import React from "react";

import { useWebRTC } from "../WebRTC";
import { useTunerStatus, Status as TunerStatus } from "../TunerStatus";
import rpc from "../rpc";

export default function Header() {
  const webRTC = useWebRTC();
  const tunerStatus = useTunerStatus();

  return (
    <header className="Header">
      <PowerButton
        active={
          tunerStatus.Connection === "Connected" &&
          tunerStatus.State === "Playing"
        }
      />
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

function PowerButton({ active }: { active: boolean }) {
  return (
    <button
      className={`PowerButton ${active ? "PowerButton--Active" : ""}`}
      onClick={() => rpc("stop").catch(console.error)}
    >
      <span hidden>Turn Off</span>
    </button>
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
    <div className="StatusIndicator">
      <div
        className={`StatusIndicator__Dot ${
          active ? "StatusIndicator__Dot--Active" : ""
        }`}
      ></div>
      <span className="StatusIndicator__Description">{description}</span>
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
