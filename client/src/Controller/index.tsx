import React from "react";

import {
  default as Backend,
  ConnectionState,
  ConnectionStatus,
} from "./Backend";

export { ConnectionStatus };

export interface State {
  connectionState: ConnectionState;
  mediaStream: undefined | MediaStream;
}

const Context = React.createContext<null | State>(null);

export const Controller = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = React.useReducer(reduce, null, () =>
    defaultState(),
  );

  React.useEffect(() => {
    const backend = new Backend();
    dispatch({ kind: "backend", backend });

    backend.on("connectionchange", (state: ConnectionState) =>
      dispatch({
        kind: "connectionchange",
        state,
      }),
    );
    backend.on("streamreceived", (stream: MediaStream) =>
      dispatch({
        kind: "streamreceived",
        stream,
      }),
    );
    backend.on("streamremoved", () => dispatch({ kind: "streamremoved" }));

    return () => {
      backend.close();
    };
  }, []);

  return <Context.Provider value={state}>{children}</Context.Provider>;
};

export const useController = (): State => {
  const state = React.useContext(Context);
  if (state === null) {
    throw new Error("useController must be called from within a <Controller>");
  }
  return state;
};

const defaultState = (): State => ({
  connectionState: { status: ConnectionStatus.Connecting },
  mediaStream: undefined,
});

type Action =
  | { kind: "backend"; backend: Backend }
  | { kind: "connectionchange"; state: ConnectionState }
  | { kind: "streamreceived"; stream: MediaStream }
  | { kind: "streamremoved" };

const reduce = (state: State, action: Action): State => {
  switch (action.kind) {
    case "backend":
      return {
        ...state,
        connectionState: action.backend.connectionState,
      };

    case "connectionchange":
      return { ...state, connectionState: action.state };

    case "streamreceived":
      return { ...state, mediaStream: action.stream };

    case "streamremoved":
      return { ...state, mediaStream: undefined };
  }
};
