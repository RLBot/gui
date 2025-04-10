import "./global.css";
import { mount } from "svelte";
import {
  BotInfo,
  HumanInfo,
  type PlayerJs,
  PsyonixBotInfo,
} from "../bindings/gui";
import App from "./App.svelte";

const app: any = mount(App, {
  target: document.body,
  // props: {
  //   name: "world",
  // },
});

export interface DraggablePlayer {
  id: string;
  displayName: string;
  icon: string;
  player: BotInfo | PsyonixBotInfo | HumanInfo;
  tags: string[];
  uniquePathSegment?: string;
}

export interface ToggleableScript {
  id: string;
  displayName: string;
  icon: string;
  config: BotInfo;
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

export function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, (c) =>
    (
      +c ^
      (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (+c / 4)))
    ).toString(16),
  );
}

export default app;
