import { EventEmitter } from "events";

export enum ConnectionStatus {
  Disconnected = "Disconnected",
  Connecting = "Connecting",
  Connected = "Connected",
  Error = "Error",
}

export type ConnectionState =
  | {
      status:
        | ConnectionStatus.Disconnected
        | ConnectionStatus.Connecting
        | ConnectionStatus.Connected;
    }
  | {
      status: ConnectionStatus.Error;
      error: Error;
    };

export enum TunerStatus {
  Stopped = "Stopped",
  Starting = "Starting",
  Playing = "Playing",
  Error = "Error",
}

export type TunerState =
  | {
      status: TunerStatus.Starting | TunerStatus.Playing;
      channelName: string;
    }
  | { status: TunerStatus.Stopped }
  | { status: TunerStatus.Error; error: Error };

type TunerStatusMessage =
  | {
      State: "Stopped";
      Error: undefined | string;
    }
  | {
      State: "Starting" | "Playing";
      ChannelName: string;
    };

type Message =
  | { Kind: "RTCOffer"; SDP: any }
  | { Kind: "ChannelList"; ChannelNames: string[] }
  | {
      Kind: "TunerStatus";
      TunerStatus: TunerStatusMessage;
    };

declare interface Backend {
  emit(event: "connectionchange", state: ConnectionState): boolean;
  on(
    event: "connectionchange",
    listener: (state: ConnectionState) => void,
  ): this;

  emit(event: "tunerchange", state: TunerState): boolean;
  on(event: "tunerchange", listener: (state: TunerState) => void): this;

  emit(event: "streamreceived", stream: MediaStream): boolean;
  on(event: "streamreceived", listener: (stream: MediaStream) => void): this;

  emit(event: "streamremoved"): boolean;
  on(event: "streamremoved", listener: () => void): this;
}

class Backend extends EventEmitter {
  private pc: RTCPeerConnection;
  private ws: WebSocket;

  private _connectionState: ConnectionState = {
    status: ConnectionStatus.Connecting,
  };
  private _mediaStream: undefined | MediaStream;

  constructor() {
    super();
    this.pc = new RTCPeerConnection();
    this.ws = new WebSocket(`ws://${window.location.host}/old-control-socket`);
    this.setup();
  }

  private setup() {
    this.ws.addEventListener("message", (evt) => this.handleSocketMessage(evt));
    this.ws.addEventListener("open", () => this.handleSocketOpen());
    this.ws.addEventListener("close", () => this.handleSocketClose());
    this.ws.addEventListener("error", (evt) => this.handleSocketError(evt));

    this.pc.addEventListener("track", (evt) =>
      this.handlePeerConnectionTrack(evt),
    );
    this.pc.addEventListener("connectionstatechange", (evt) =>
      console.log("Connection state", this.pc.connectionState, evt),
    );
    this.pc.addEventListener("signalingstatechange", (evt) =>
      console.log("Signaling state", this.pc.signalingState, evt),
    );
  }

  get connectionState() {
    return this._connectionState;
  }

  close() {
    this.pc.close();
    this.ws.close();
  }

  private handleSocketMessage(evt: MessageEvent) {
    const message: Message = JSON.parse(evt.data);
    console.log("Received message", message);
    switch (message.Kind) {
      case "RTCOffer":
        this.handleRTCOffer(message.SDP);
        break;

      case "TunerStatus":
        this.handleTunerStatus(message.TunerStatus);
        break;

      default:
        break;
    }
  }

  private async handleRTCOffer(sdp: any) {
    console.log("Received remote description", sdp);
    this.pc.setRemoteDescription(sdp);

    const answer = await this.pc.createAnswer();
    console.log("Created local description", answer);
    await this.pc.setLocalDescription(answer);

    this.ws.send(
      JSON.stringify({
        Kind: "RTCAnswer",
        SDP: answer,
      }),
    );
  }

  private handleTunerStatus(status: TunerStatusMessage) {
    switch (status.State) {
      case "Stopped":
        if (status.Error) {
          this.emit("tunerchange", {
            status: TunerStatus.Error,
            error: new Error(status.Error),
          });
        } else {
          this.emit("tunerchange", {
            status: TunerStatus.Stopped,
          });
        }
        break;

      case "Starting":
        this.emit("tunerchange", {
          status: TunerStatus.Starting,
          channelName: status.ChannelName,
        });
        break;

      case "Playing":
        this.emit("tunerchange", {
          status: TunerStatus.Playing,
          channelName: status.ChannelName,
        });
        break;
    }
  }

  private handleSocketOpen() {
    console.log("WebSocket opened");
    this._connectionState = { status: ConnectionStatus.Connected };
    this.emit("connectionchange", this._connectionState);
  }

  private handleSocketClose() {
    console.log("WebSocket closed");
    this._connectionState = { status: ConnectionStatus.Disconnected };
    this.emit("connectionchange", this._connectionState);
  }

  private handleSocketError(evt: Event) {
    console.log("WebSocket error", evt);
    this._connectionState = {
      status: ConnectionStatus.Error,
      error: new Error(evt.toString()),
    };
    this.emit("connectionchange", this._connectionState);
  }

  private handlePeerConnectionTrack(evt: RTCTrackEvent) {
    console.log("RTC track", evt);

    if (evt.streams.length < 1) {
      return;
    }

    const stream = evt.streams[0];
    if (this._mediaStream && this._mediaStream.id === stream.id) {
      return;
    }

    this._mediaStream = stream;
    stream.addEventListener("removetrack", () =>
      this.handleMediaStreamRemoveTrack(stream),
    );

    this.emit("streamreceived", stream);
  }

  private handleMediaStreamRemoveTrack(stream: MediaStream) {
    console.log("Track removed from stream", stream);

    if (!this._mediaStream || this._mediaStream.id !== stream.id) {
      return;
    }

    this._mediaStream = undefined;
    this.emit("streamremoved");
  }
}

export default Backend;
