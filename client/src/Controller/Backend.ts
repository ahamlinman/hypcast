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
  Started = "Started",
  Stopped = "Stopped",
  Error = "Error",
}

export type TunerState =
  | {
      status: TunerStatus.Started;
      channelName: string;
    }
  | { status: TunerStatus.Stopped }
  | { status: TunerStatus.Error; error: Error };

type TunerStatusMessage = {
  ChannelName: undefined | string;
  Error: undefined | string;
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

  emit(event: "channellistreceived", channelList: string[]): boolean;
  on(
    event: "channellistreceived",
    listener: (channelList: string[]) => void,
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
    this.ws = new WebSocket(`ws://${window.location.host}/control-socket`);
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
  }

  get connectionState() {
    return this._connectionState;
  }

  close() {
    this.pc.close();
    this.ws.close();
  }

  changeChannel(channelName: string) {
    this.ws.send(
      JSON.stringify({
        Kind: "ChangeChannel",
        ChannelName: channelName,
      }),
    );
  }

  turnOff() {
    this.ws.send(
      JSON.stringify({
        Kind: "TurnOff",
      }),
    );
  }

  private handleSocketMessage(evt: MessageEvent) {
    const message: Message = JSON.parse(evt.data);
    switch (message.Kind) {
      case "RTCOffer":
        this.handleRTCOffer(message.SDP);
        break;

      case "ChannelList":
        this.emit("channellistreceived", message.ChannelNames);
        break;

      case "TunerStatus":
        this.handleTunerStatus(message.TunerStatus);
        break;

      default:
        break;
    }
  }

  private async handleRTCOffer(sdp: any) {
    this.pc.setRemoteDescription(sdp);

    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);

    this.ws.send(
      JSON.stringify({
        Kind: "RTCAnswer",
        SDP: answer,
      }),
    );
  }

  private handleTunerStatus(status: TunerStatusMessage) {
    if (status.Error) {
      this.emit("tunerchange", {
        status: TunerStatus.Error,
        error: new Error(status.Error),
      });
    } else if (status.ChannelName) {
      this.emit("tunerchange", {
        status: TunerStatus.Started,
        channelName: status.ChannelName,
      });
    } else {
      this.emit("tunerchange", {
        status: TunerStatus.Stopped,
      });
    }
  }

  private handleSocketOpen() {
    this._connectionState = { status: ConnectionStatus.Connected };
    this.emit("connectionchange", this._connectionState);
  }

  private handleSocketClose() {
    this._connectionState = { status: ConnectionStatus.Disconnected };
    this.emit("connectionchange", this._connectionState);
  }

  private handleSocketError(evt: Event) {
    this._connectionState = {
      status: ConnectionStatus.Error,
      error: new Error(evt.toString()),
    };
    this.emit("connectionchange", this._connectionState);
  }

  private handlePeerConnectionTrack(evt: RTCTrackEvent) {
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
    if (!this._mediaStream || this._mediaStream.id !== stream.id) {
      return;
    }

    this._mediaStream = undefined;
    this.emit("streamremoved");
  }
}

export default Backend;
