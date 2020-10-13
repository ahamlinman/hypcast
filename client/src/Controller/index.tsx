import React from "react";

import {
  default as Backend,
  ConnectionState,
  ConnectionStatus,
  TunerState,
  TunerStatus,
} from "./Backend";

export { ConnectionStatus, TunerStatus };

export interface State {
  connectionState: ConnectionState;

  tunerState: undefined | TunerState;
  mediaStream: undefined | MediaStream;

  changeChannel: (channelName: string) => void;
  currentChannelName: undefined | string;

  turnOff: () => void;
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
    backend.on("tunerchange", (state: TunerState) =>
      dispatch({
        kind: "tunerchange",
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
  tunerState: undefined,
  mediaStream: undefined,
  changeChannel: () => {},
  currentChannelName: undefined,
  turnOff: () => {},
});

type Action =
  | { kind: "backend"; backend: Backend }
  | { kind: "connectionchange"; state: ConnectionState }
  | { kind: "tunerchange"; state: TunerState }
  | { kind: "streamreceived"; stream: MediaStream }
  | { kind: "streamremoved" }
  | { kind: "changechannel"; channelName: string };

const reduce = (state: State, action: Action): State => {
  switch (action.kind) {
    case "backend":
      return {
        ...state,
        connectionState: action.backend.connectionState,
        changeChannel: action.backend.changeChannel.bind(action.backend),
        turnOff: action.backend.turnOff.bind(action.backend),
      };

    case "connectionchange":
      return { ...state, connectionState: action.state };

    case "tunerchange":
      return {
        ...state,
        tunerState: action.state,
        currentChannelName:
          action.state.status === TunerStatus.Started
            ? action.state.channelName
            : undefined,
      };

    case "streamreceived":
      return { ...state, mediaStream: action.stream };

    case "streamremoved":
      return { ...state, mediaStream: undefined };
  }

  return state;
};
