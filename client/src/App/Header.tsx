import React from "react";

import { useWebRTC, State as WebRTCState } from "../WebRTC";
import { useTunerStatus, Status as TunerStatus } from "../TunerStatus";
import rpc from "../rpc";

export default function Header() {
  return (
    <header className="Header">
      <PowerButton />
      <Title />
      <StatusIndicator />
    </header>
  );
}

function Title() {
  return <h1>hypcast</h1>;
}

function PowerButton() {
  const tunerStatus = useTunerStatus();
  const active = isTunerActive(tunerStatus);

  return (
    <button
      className={`PowerButton ${active ? "PowerButton--Active" : ""}`}
      onClick={() => rpc("stop").catch(console.error)}
      aria-label="Turn Off"
    />
  );
}

function StatusIndicator() {
  const webRTC = useWebRTC();
  const tunerStatus = useTunerStatus();

  const active = isActive(webRTC, tunerStatus);

  return (
    <div className="StatusIndicator">
      <div
        className={`StatusIndicator__Dot ${
          active ? "StatusIndicator__Dot--Active" : ""
        }`}
      ></div>
      <span className="StatusIndicator__Description">
        {statusString(webRTC, tunerStatus)}
      </span>
    </div>
  );
}

function statusString(webRTC: WebRTCState, tunerStatus: TunerStatus): string {
  const webRTCString = webRTCStatusString(webRTC);
  if (webRTCString !== undefined) {
    return webRTCString;
  }

  return tunerStatusString(tunerStatus);
}

function webRTCStatusString(state: WebRTCState): string | undefined {
  if (state.Connection.Status !== "Connected") {
    return state.Connection.Status;
  }

  return undefined;
}

function tunerStatusString(status: TunerStatus) {
  if (status.Connection !== "Connected") {
    return status.Connection;
  }

  if (status.State === "Stopped") {
    return status.State;
  }

  return `${status.State} ${status.ChannelName}`;
}

function isActive(webRTC: WebRTCState, tunerStatus: TunerStatus): boolean {
  if (webRTC.Connection.Status !== "Connected") {
    return false;
  }

  return isTunerActive(tunerStatus);
}

function isTunerActive(status: TunerStatus) {
  return status.Connection === "Connected" && status.State === "Playing";
}
