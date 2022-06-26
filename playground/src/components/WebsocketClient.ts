import { multiplayer } from "../generated/multiplayer/v1/multiplayer";

const MESSAGE_TYPES = {
  UNSPECIFIED: 0,
  UPDATE_PLAYER_DATA: 1,
  UPDATE_GAME_STATE: 2,
  SET_CLIENT_ID: 3,
  DISCONNECT_CLIENT: 4,
  CLIENT_ADDED: 5,
};

const MESSAGE_TYPES_VALUES = {
  UNSPECIFIED: "UNSPECIFIED",
  UPDATE_PLAYER_DATA: "UPDATE_PLAYER_DATA",
  UPDATE_GAME_STATE: "UPDATE_GAME_STATE",
  SET_CLIENT_ID: "SET_CLIENT_ID",
  DISCONNECT_CLIENT: "DISCONNECT_CLIENT",
  CLIENT_ADDED: "CLIENT_ADDED",
};

export interface IPlayerData {
  name: string;
  x: number;
  y: number;
  color: string;
}

const createUpdatePlayerDataMessage = (data: IPlayerData) => {
  const msg = new multiplayer.v1.Message({
    type: multiplayer.v1.MessageType.UPDATE_PLAYER_DATA,
    data: JSON.stringify(data),
  });

  return JSON.stringify(msg.toObject());
};

export class WebsocketClient {
  private websocket: WebSocket;

  constructor(url: string) {
    this.websocket = new WebSocket(url);

    this.websocket.onopen = (e: Event) => {
      this._onConnect();
    };

    this.websocket.onmessage = (e: MessageEvent<any>) => {
      if (e.data instanceof Blob) {
        e.data.text().then(this.handleMessage);
      } else {
        this.handleMessage(e.data);
      }
    };
  }

  handleMessage = (data: any) => {
    const msg = JSON.parse(data);
    switch (msg.type) {
      case "UPDATE_GAME_STATE":
        this._onUpdateGameState(JSON.parse(msg.data));
        break;
      case "SET_CLIENT_ID":
        this._onSetClientId(msg.data);
        break;
      default:
        console.error("no handler for message type ", msg.type);
    }
  };

  private send = (msg: string) => {
    const bytes = new TextEncoder().encode(msg);
    const blob = new Blob([bytes]);
    this.websocket.send(blob);
  };

  public updatePlayerState = (data: IPlayerData) => {
    const msg = createUpdatePlayerDataMessage(data);
    this.send(msg);
  };

  private _onConnect: () => void = () => console.log("onConnect not implemented");
  private _onUpdateGameState: (data: any) => void = () => console.log("onUpdateGameState not implemented");
  private _onSetClientId: (data: any) => void = () => console.log("onSetClientId not implemented");

  public onConnect: RegisterOnConnectHandler = (fn) => {
    this._onConnect = fn;
    return this;
  };
  public onUpdateGameState: RegisterCallBackHandler<any> = (fn) => {
    this._onUpdateGameState = fn;
    return this;
  };
  public onSetClientId: RegisterCallBackHandler<any> = (fn) => {
    this._onSetClientId = fn;
    return this;
  };
}

type RegisterOnConnectHandler = (fn: () => void) => WebsocketClient;
type RegisterCallBackHandler<T> = (fn: (data: T) => void) => WebsocketClient;
