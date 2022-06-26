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
  return JSON.stringify({
    type: MESSAGE_TYPES.UPDATE_PLAYER_DATA,
    data: JSON.stringify(data),
  });
};

export class WebsocketClient {
  private websocket: WebSocket;

  constructor() {
    this.websocket = new WebSocket("ws://localhost:3000/wsb");

    this.websocket.onopen = (e: Event) => {
      //   console.log("opened connection", e);
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
    // console.log(data);
  };

  public sendBinary = (msg: string) => {
    const bytes = new TextEncoder().encode(msg);
    const blob = new Blob([bytes]);
    this.websocket.send(blob);
  };

  public update = (data: IPlayerData) => {
    const msg = createUpdatePlayerDataMessage(data);
    this.sendBinary(msg);
  };
}
