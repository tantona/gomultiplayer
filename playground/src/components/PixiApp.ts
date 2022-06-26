import * as PIXI from "pixi.js";
import { WebsocketClient } from "./WebsocketClient";
import randomcolor from "randomcolor";
import { faker } from "@faker-js/faker";

type IPlayer = { id?: string; name: string; x: number; y: number; color: string };
type IPlayers = Record<string, IPlayer>;
interface IState {
  // id: string | null;
  // position: { x: number; y: number };

  me: IPlayer;
  players: IPlayers;
}

class Player {
  private state: IPlayer;

  public container = new PIXI.Container();
  private namePlate = new PIXI.Text("", {
    fill: "#FFFFFF",
    stroke: "#FFFFFF",
    fontSize: 12,
  });

  private cursor = new PIXI.Graphics();
  private size = 10;
  constructor(state: IPlayer) {
    this.state = state;
    this.namePlate.text = this.state.name;
    this.namePlate.x = (this.namePlate.width / 2) * -1;
    this.namePlate.y = -25;
    this.cursor.beginFill(toBytes(this.state.color)).drawCircle(0, 0, this.size).endFill();
    this.container.addChild(this.cursor).addChild(this.namePlate);
  }

  public setPosition(pos: { x: number; y: number }) {
    this.container.x = pos.x;
    this.container.y = pos.y;
  }

  public draw() {
    this.cursor.beginFill(toBytes(this.state.color)).drawCircle(0, 0, 10).endFill();
  }
}

export class PixiApp {
  private app = new PIXI.Application({ width: 600, height: 600, backgroundAlpha: 1, antialias: true });
  private websocket = new WebsocketClient("ws://localhost:3000/wsb");

  private state: IState = {
    me: {
      id: "",
      name: faker.internet.userName(),
      x: 0,
      y: 0,
      color: randomcolor(),
    },
    players: {},
  };

  private me = new Player({
    name: this.state.me.name,
    x: 0,
    y: 0,
    color: this.state.me.color,
  });

  private background = new PIXI.Graphics()
    .beginFill(0x000000)
    .drawRect(0, 0, 600, 600)
    .endFill()
    .on("pointermove", (e) => {
      this.state.me.x = e.data.global.x;
      this.state.me.y = e.data.global.y;

      this.websocket.updatePlayerState({
        name: this.state.me.name,
        color: this.state.me.color.toString(),
        x: this.state.me.x,
        y: this.state.me.y,
      });
    });

  public get view() {
    return this.app.view;
  }

  public initialize = () => {
    this.websocket.onConnect(() =>
      this.websocket.updatePlayerState({
        name: this.state.me.name,
        x: this.state.me.x,
        y: this.state.me.y,
        color: this.state.me.color,
      })
    );
    this.websocket.onSetClientId((data) => (this.state.me.id = data));
    this.websocket.onUpdateGameState((data) => (this.state.players = data?.players));

    this.app.renderer.plugins.interaction["moveWhenInside"] = true;

    this.background.interactive = true;

    this.app.stage.addChild(this.background);
    this.app.ticker.add(() => this.update());
  };

  private players: Record<string, Player> = {};

  private drawPlayers() {
    for (const id of Object.keys(this.state.players)) {
      if (id === this.state.me.id) {
        continue;
      }
      const player = this.state.players[id];

      let p = this.players[id];
      if (!p) {
        this.players[id] = new Player(player);
        p = this.players[id];
      }

      p.setPosition({
        x: player.x,
        y: player.y,
      });

      this.background.addChild(p.container);
    }
  }

  private drawMe() {
    this.me.setPosition({
      x: this.state.me.x,
      y: this.state.me.y,
    });
    this.background.addChild(this.me.container);
  }

  private update = () => {
    this.drawMe();
    this.drawPlayers();
  };
}

const toBytes = (color: string) => parseInt(color.replace(/^#/, ""), 16);
