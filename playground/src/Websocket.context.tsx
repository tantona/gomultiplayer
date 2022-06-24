import { FC, Dispatch, createContext, Reducer, useReducer, PropsWithChildren, useContext } from "react";
import useWebSocket from "react-use-websocket";

export interface IState {
  clientId?: string | null;
  game?: any | null;
}

type UpdateGameState = {
  type: "UPDATE_GAME_STATE";
  data: any;
};

type SetClientId = {
  type: "SET_CLIENT_ID";
  data: string;
};

export interface IPlayerData {
  name: string;
  x: number;
  y: number;
  color: string;
}

const INITIAL_STATE: IState = {
  clientId: null,
  game: null,
};

const ctx = createContext<{
  state: IState;
  dispatch: Dispatch<DispatcherAction>;
  update: (_data: IPlayerData) => void;
}>({
  state: INITIAL_STATE,
  dispatch: () => null,
  update: () => null,
});

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

const getColorCode = () => {
  const makeColorCode = "0123456789ABCDEF";
  let code = "#";
  for (let count = 0; count < 6; count++) {
    code = code + makeColorCode[Math.floor(Math.random() * 16)];
  }
  return code;
};

type DispatcherAction = UpdateGameState | SetClientId;

const createUpdatePlayerDataMessage = (data: IPlayerData) => {
  return JSON.stringify({
    type: MESSAGE_TYPES.UPDATE_PLAYER_DATA,
    data: JSON.stringify(data),
  });
};

export const Provider: FC<PropsWithChildren<{ path: string }>> = ({ path = "ws://localhost:8080/wsb", children }) => {
  const mycolor = "#533a71";
  const { sendMessage, lastMessage, readyState } = useWebSocket(path, {
    onOpen: () => {
      update({ name: "", x: 0, y: 0, color: mycolor });
    },
    onMessage: (e) => {
      if (e.data instanceof Blob) {
        e.data.text().then(messageHandler);
      } else {
        messageHandler(e.data);
      }
    },
  });

  const [state, dispatch] = useReducer<Reducer<IState, DispatcherAction>>((s, action) => {
    switch (action.type) {
      case "SET_CLIENT_ID":
        return { ...s, clientId: action.data };
      case "UPDATE_GAME_STATE":
        return { ...s, game: action.data };
      default:
        return s;
    }
  }, INITIAL_STATE);

  const sendBinary = (msg: string) => {
    const bytes = new TextEncoder().encode(msg);
    const blob = new Blob([bytes]);

    sendMessage(blob);
  };

  const messageHandler = (data: any) => {
    const msg = JSON.parse(data);

    switch (msg.type) {
      case MESSAGE_TYPES_VALUES.SET_CLIENT_ID:
        dispatch({ type: "SET_CLIENT_ID", data: msg.data });
        break;
      case MESSAGE_TYPES_VALUES.UPDATE_GAME_STATE:
        dispatch({ type: "UPDATE_GAME_STATE", data: JSON.parse(msg.data) });
        break;
      default:
        break;
    }
  };

  const update = (data: IPlayerData) => {
    const msg = createUpdatePlayerDataMessage(data);
    sendBinary(msg);
  };

  return <ctx.Provider value={{ state, dispatch, update }}>{children}</ctx.Provider>;
};

export const useWebsocket = () => {
  return useContext(ctx);
};
