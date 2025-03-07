import "./global.css";
import App from "./App.svelte";
import { mount } from "svelte";
import {
  BotInfo,
  type PlayerJs,
  PsyonixBotInfo,
  HumanInfo,
} from "../bindings/gui";

const app: any = mount(App, {
  target: document.body,
  // props: {
  //   name: "world",
  // },
});

export interface DraggablePlayer {
  id: number;
  displayName: string;
  icon: string;
  player: BotInfo | PsyonixBotInfo | HumanInfo;
  tags: string[];
  uniquePathSegment?: string;
}

export function draggablePlayerToPlayerJs(d: DraggablePlayer): PlayerJs {
  let sort = "";

  if (d.player instanceof BotInfo) {
    sort = "rlbot";
  }
  if (d.player instanceof PsyonixBotInfo) {
    sort = "psyonix";
  }
  if (d.player instanceof HumanInfo) {
    sort = "human";
  }

  return {
    sort: sort,
    player: d.player,
  };
}

export default app;
