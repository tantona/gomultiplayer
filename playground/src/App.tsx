// import { Stage, Container } from "@inlet/react-pixi";
// import * as pixi from "pixi.js";
import { useInterval } from "react-use";
import { useEffect, useRef } from "react";
import { IPlayerData, IState, Provider, useWebsocket } from "./Websocket.context";

const drawPlayer = (ctx: CanvasRenderingContext2D, player: IPlayerData) => {
  ctx.fillStyle = player?.color ?? "#000";
  ctx.fillRect(player?.x ?? 0, player?.y ?? 0, 10, 10);
};

const drawMe = (ctx: CanvasRenderingContext2D, state: IState) => {
  const clientId = state?.clientId;
  if (clientId) {
    const me = state?.game?.players?.[clientId];
    drawPlayer(ctx, me);
  }
};

const drawOthers = (ctx: CanvasRenderingContext2D, state: IState) => {
  const clientId = state?.clientId;
  Object.keys(state?.game?.players ?? {}).forEach((id) => {
    if (id !== clientId) {
      drawPlayer(ctx, state?.game?.players[id]);
    }
  });
};

const draw = (ctx: CanvasRenderingContext2D, state: IState) => {
  ctx.fillStyle = "#ba1b1b";
  ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height);
  drawMe(ctx, state);
  drawOthers(ctx, state);
};

const App = () => {
  const { update, state } = useWebsocket();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    canvasRef?.current?.addEventListener(
      "mousemove",
      (e) => {
        update({ name: "", color: "#533a71", x: e.offsetX, y: e.offsetY });
      },
      false
    );
  }, []);

  useInterval(() => {
    const canvas = canvasRef.current;
    const context = canvas?.getContext("2d");
    if (context) {
      draw(context, state);
    }
  }, 16);

  return <canvas ref={canvasRef} width="500" height="500" style={{ border: "1px solid black" }}></canvas>;
};

export const AppContainer = () => {
  return (
    <Provider path="ws://localhost:3000/wsb">
      <App />
    </Provider>
  );
};
