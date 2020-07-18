import React from "react";

export enum Status {
  Closed = "Closed",
  Connecting = "Connecting",
  Connected = "Connected",
  Error = "Error",
}

export interface Controller {
  status: Status;
  audioDuration: number;
  videoDuration: number;
}

const defaultController = (): Controller => ({
  status: Status.Closed,
  audioDuration: 0,
  videoDuration: 0,
});

interface EventData {
  Kind: string;
  Duration: number;
}

export const Context = React.createContext<Controller>(defaultController());

const Controller = ({ children }: { children: React.ReactNode }) => {
  const [controller, setController] = React.useState(defaultController());

  React.useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.host}/hypcast/ws`);

    ws.addEventListener("open", (evt) => {
      setController((c) => ({ ...c, status: Status.Connected }));
    });

    ws.addEventListener("error", (evt) => {
      setController((c) => ({ ...c, status: Status.Error }));
    });

    ws.addEventListener("close", (evt) => {
      setController((c) => ({
        ...c,
        status: c.status !== Status.Error ? Status.Closed : Status.Error,
      }));
    });

    ws.addEventListener("message", (evt) => {
      const data: EventData = JSON.parse(evt.data);
      switch (data.Kind) {
        case "audio":
          setController((c) => ({
            ...c,
            audioDuration: c.audioDuration + data.Duration,
          }));
          break;
        case "video":
          setController((c) => ({
            ...c,
            videoDuration: c.videoDuration + data.Duration,
          }));
          break;
      }
    });

    return () => {
      ws.close();
    };
  }, []);

  return <Context.Provider value={controller}>{children}</Context.Provider>;
};

export default Controller;
