import * as PIXI from "pixi.js";
import { WebsocketClient } from "./WebsocketClient";

const getColorCode = () => {
  const makeColorCode = "0123456789ABCDEF";
  let code = "#";
  for (let count = 0; count < 6; count++) {
    code = code + makeColorCode[Math.floor(Math.random() * 16)];
  }
  return code;
};

export class PixiApp {
  private app = new PIXI.Application({ width: 600, height: 600, backgroundAlpha: 1, antialias: true });
  private websocket = new WebsocketClient();
  private state = {
    position: {
      x: 0,
      y: 0,
    },
  };
  private mycolor = getColorCode();
  private cursor = new PIXI.Graphics().beginFill(0x533a71).drawCircle(0, 0, 10).endFill();
  private background = new PIXI.Graphics()
    .beginFill(0x000000)
    .drawRect(0, 0, 600, 600)
    .endFill()
    .on("pointermove", (e) => {
      this.state.position.x = e.data.global.x;
      this.state.position.y = e.data.global.y;

      this.websocket.update({
        name: "foobar",
        color: this.mycolor,
        ...this.state.position,
      });
    });

  public get view() {
    return this.app.view;
  }

  public initialize = () => {
    this.app.renderer.plugins.interaction["moveWhenInside"] = true;

    this.background.interactive = true;

    this.app.stage.addChild(this.background);
    this.app.ticker.add(() => this.update());
  };

  private update = () => {
    this.cursor.x = this.state.position.x;
    this.cursor.y = this.state.position.y;
    this.background.addChild(this.cursor);
  };
}
