import React from "react";

import { default as Backend, ConnectionState } from "./Backend";

export interface State {
  Connection: ConnectionState;
  MediaStream: undefined | MediaStream;
}

const Context = React.createContext<null | State>(null);

export const useWebRTC = (): State => {
  const state = React.useContext(Context);
  if (state === null) {
    throw new Error("useWebRTC must be used within <WebRTCProvider>");
  }
  return state;
};

export const WebRTCProvider = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = React.useReducer(reduce, null, () =>
    defaultState(),
  );

  React.useEffect(() => {
    const backend = new Backend();
    dispatch({ kind: "connectionchange", state: backend.connectionState });

    backend.on("connectionchange", (state: ConnectionState) =>
      dispatch({ kind: "connectionchange", state }),
    );
    backend.on("streamreceived", (stream: MediaStream) =>
      dispatch({ kind: "streamreceived", stream }),
    );
    backend.on("streamremoved", () => dispatch({ kind: "streamremoved" }));

    return () => {
      backend.close();
    };
  }, []);

  return <Context.Provider value={state}>{children}</Context.Provider>;
};

const defaultState = (): State => ({
  Connection: { Status: "Connecting" },
  MediaStream: undefined,
});

type Action =
  | { kind: "connectionchange"; state: ConnectionState }
  | { kind: "streamreceived"; stream: MediaStream }
  | { kind: "streamremoved" };

const reduce = (state: State, action: Action): State => {
  switch (action.kind) {
    case "connectionchange":
      return { ...state, Connection: action.state };

    case "streamreceived":
      return { ...state, MediaStream: action.stream };

    case "streamremoved":
      return { ...state, MediaStream: undefined };
  }
};
