import React from "react";

type TunerStatus =
  | { State: "Starting" | "Playing"; ChannelName: string }
  | { State: "Stopped"; Error: undefined | string };

export type Status =
  | { Connection: "Disconnected" | "Connecting" }
  | ({ Connection: "Connected" } & TunerStatus);

const Context = React.createContext<null | Status>(null);

export const useTunerStatus = (): Status => {
  const status = React.useContext(Context);
  if (status === null) {
    throw new Error("useTunerStatus must be used within <TunerStatusProvider>");
  }
  return status;
};

export const TunerStatusProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [status, setStatus] = React.useState<Status>({
    Connection: "Connecting",
  });

  React.useEffect(() => {
    const ws = new WebSocket(
      `ws://${window.location.host}/api/socket/tuner-status`,
    );

    let closed = false;
    const close = () => {
      if (closed) {
        return;
      }
      closed = true;
      ws.onmessage = null;
      ws.onclose = null;
      ws.onerror = null;
      ws.close();
    };

    ws.onmessage = (evt) => {
      const status: TunerStatus = JSON.parse(evt.data);
      console.log("Received tuner status", status);
      setStatus({ Connection: "Connected", ...status });
    };

    ws.onclose = () => {
      console.log("Tuner status socket closed");
      setStatus({ Connection: "Disconnected" });
      close();
    };
    ws.onerror = (evt) => {
      console.error("Tuner status socket error", evt);
      setStatus({ Connection: "Disconnected" });
      close();
    };

    return close;
  }, []);

  return <Context.Provider value={status}>{children}</Context.Provider>;
};
