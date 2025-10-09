import "./global.css";
import { mount } from "svelte";
import {
  BotInfo,
  HumanInfo,
  type PlayerJs,
  PsyonixBotInfo,
} from "../bindings/gui";
import App from "./App.svelte";
import SuperJSON from "superjson";

const app: any = mount(App, {
  target: document.body,
  // props: {
  //   name: "world",
  // },
});

SuperJSON.registerClass(BotInfo);
SuperJSON.registerClass(PsyonixBotInfo);
SuperJSON.registerClass(HumanInfo);

export function parseJSON(item: string | null): any | null {
  if (item === null) {
    return null;
  }

  try {
    return JSON.parse(item);
  } catch {
    return null;
  }
}

export interface DraggablePlayer {
  id: string;
  displayName: string;
  icon: string;
  player: BotInfo | PsyonixBotInfo | HumanInfo;
  tags: string[];
  uniquePathSegment?: string;
  modified?: boolean;
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

export default app;
