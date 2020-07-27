import React from "react";

export interface State {
  connected: boolean;
  stream: null | MediaStream;
}

const Context = React.createContext<null | [State, React.Dispatch<Action>]>(
  null,
);

export const Controller = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = React.useReducer(reduce, null, () =>
    defaultState(),
  );

  React.useEffect(() => {
    const pc = new RTCPeerConnection();
    const ws = new WebSocket(`ws://${window.location.host}/control-socket`);

    pc.addEventListener("track", (evt) => {
      console.log("Received new track:", evt.track, evt.streams);

      if (evt.track.kind === "audio") {
        return;
      }

      dispatch({ kind: "ReceivedStream", stream: evt.streams[0] });
    });

    ws.addEventListener("open", () => {
      dispatch({ kind: "Connected" });
    });

    ws.addEventListener("close", () => {
      dispatch({ kind: "Disconnected" });
    });

    ws.addEventListener("message", (evt) => {
      const message = JSON.parse(evt.data);
      switch (message.Kind) {
        case "RTCOffer":
          console.log("Received offer message", message);
          (async () => {
            pc.setRemoteDescription(message.SDP);

            const answer = await pc.createAnswer();
            await pc.setLocalDescription(answer);

            const response = {
              Kind: "RTCAnswer",
              SDP: answer,
            };

            console.log("Sending response message:", response);
            ws.send(JSON.stringify(response));
          })();
          break;

        default:
          console.error("Unrecognized message", message);
          break;
      }
    });

    return () => {
      dispatch({ kind: "Disconnected" });
      ws.close();
      pc.close();
    };
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
  connected: false,
  stream: null,
});

type Action = InternalAction | ExternalAction;

export type ExternalAction = never;

type InternalAction =
  | ActionConnected
  | ActionDisconnected
  | ActionReceivedStream;

interface ActionConnected {
  kind: "Connected";
}

interface ActionDisconnected {
  kind: "Disconnected";
}

interface ActionReceivedStream {
  kind: "ReceivedStream";
  stream: MediaStream;
}

const reduce = (state: State, action: Action) => {
  switch (action.kind) {
    case "Connected":
      return { ...state, connected: true };

    case "Disconnected":
      return { ...state, connected: false, stream: null };

    case "ReceivedStream":
      return { ...state, stream: action.stream };

    default:
      return state;
  }
};
