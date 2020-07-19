import React from "react";

export enum SocketStatus {
  Closed = "Closed",
  Connected = "Connected",
}

export interface State {
  socketStatus: SocketStatus;
  audioDuration: number;
  videoDuration: number;
}

export const Context = React.createContext<
  null | [State, React.Dispatch<Action>]
>(null);

export const Controller = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = React.useReducer(reduce, null, () =>
    defaultState(),
  );

  React.useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.host}/hypcast/ws`);

    ws.addEventListener("open", () => {
      dispatch({ kind: "UpdateSocketStatus", status: SocketStatus.Connected });
    });

    ws.addEventListener("close", () => {
      dispatch({ kind: "UpdateSocketStatus", status: SocketStatus.Closed });
    });

    ws.addEventListener("message", (evt) => {
      interface EventData {
        Kind: string;
        Duration: number;
      }

      const data: EventData = JSON.parse(evt.data);
      switch (data.Kind) {
        case "audio":
          dispatch({ kind: "IncrementAudioDuration", duration: data.Duration });
          break;

        case "video":
          dispatch({ kind: "IncrementVideoDuration", duration: data.Duration });
          break;

        default:
          console.error("Unrecognized event:", data);
          break;
      }
    });
  }, []);

  return (
    <Context.Provider value={[state, dispatch]}>{children}</Context.Provider>
  );
};

export const useController = (): [State, React.Dispatch<ExternalAction>] => {
  const ctx = React.useContext(Context);
  if (ctx === null) {
    throw new Error(
      "useController must only be used by children of a <Controller>",
    );
  }

  const [state, dispatch] = ctx;
  return [state, (action: ExternalAction) => dispatch(action)];
};

const defaultState = () => ({
  socketStatus: SocketStatus.Closed,
  audioDuration: 0,
  videoDuration: 0,
});

type Action = InternalAction | ExternalAction;

type ExternalAction = never;

type InternalAction =
  | ActionUpdateSocketStatus
  | ActionIncrementAudioDuration
  | ActionIncrementVideoDuration;

interface ActionUpdateSocketStatus {
  kind: "UpdateSocketStatus";
  status: SocketStatus;
}

interface ActionIncrementAudioDuration {
  kind: "IncrementAudioDuration";
  duration: number;
}

interface ActionIncrementVideoDuration {
  kind: "IncrementVideoDuration";
  duration: number;
}

const reduce = (state: State, action: Action) => {
  switch (action.kind) {
    case "UpdateSocketStatus":
      return { ...state, socketStatus: action.status };

    case "IncrementAudioDuration":
      return { ...state, audioDuration: state.audioDuration + action.duration };

    case "IncrementVideoDuration":
      return { ...state, videoDuration: state.videoDuration + action.duration };

    default:
      return state;
  }
};
