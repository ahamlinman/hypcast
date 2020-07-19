import React from "react";

export interface State {
  connected: boolean;
  stream: null | MediaStream;
}

export const Context = React.createContext<
  null | [State, React.Dispatch<Action>]
>(null);

export const Controller = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = React.useReducer(reduce, null, () =>
    defaultState(),
  );

  React.useEffect(() => {
    const pc = new RTCPeerConnection();
    pc.addTransceiver("video", { direction: "sendrecv" });
    pc.addTransceiver("audio", { direction: "sendrecv" });

    pc.addEventListener("track", (evt) => {
      console.log("Received new track:", evt.track, evt.streams);

      if (evt.track.kind === "audio") {
        return;
      }

      dispatch({ kind: "SetStream", value: evt.streams[0] });
    });

    const ws = new WebSocket(`ws://${window.location.host}/hypcast/ws`);

    ws.addEventListener("open", () => {
      dispatch({ kind: "SetConnected", value: true });
    });

    ws.addEventListener("close", () => {
      dispatch({ kind: "SetConnected", value: false });
      dispatch({ kind: "SetStream", value: null });
    });

    ws.addEventListener("message", (evt) => {
      const message = JSON.parse(evt.data);
      switch (message.Kind) {
        case "ServerOffer":
          console.log("Received offer message", message);
          (async () => {
            pc.setRemoteDescription(message.ServerOffer);

            const answer = await pc.createAnswer();
            await pc.setLocalDescription(answer);

            const response = {
              Kind: "ClientAnswer",
              ClientAnswer: answer,
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
      dispatch({ kind: "SetStream", value: null });
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

type ExternalAction = never;

type InternalAction = ActionSetConnected | ActionSetStream;

interface ActionSetConnected {
  kind: "SetConnected";
  value: boolean;
}

interface ActionSetStream {
  kind: "SetStream";
  value: null | MediaStream;
}

const reduce = (state: State, action: Action) => {
  switch (action.kind) {
    case "SetConnected":
      return { ...state, connected: action.value };

    case "SetStream":
      return { ...state, stream: action.value };

    default:
      return state;
  }
};
