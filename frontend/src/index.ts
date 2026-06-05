import "./global.css";
import { mount } from "svelte";
import {
  BotInfo,
  HumanInfo,
  LoadoutConfig,
  type PlayerJs,
  PsyonixBotInfo,
} from "../bindings/gui";
import App from "./App.svelte";
import SuperJSON from "superjson";

SuperJSON.registerClass(BotInfo);
SuperJSON.registerClass(PsyonixBotInfo);
SuperJSON.registerClass(HumanInfo);
SuperJSON.registerClass(LoadoutConfig);

const app = mount(App, {
  target: document.body,
  // props: {
  //   name: "world",
  // },
});

export function parseJSON(item: string | null): any | null {
  if (item === null) {
    return null;
  }

  try {
    return JSON.parse(item);
  } catch (e) {
    console.warn("JSON Parse error", e);
    return null;
  }
}

export function parseSuperJSON(item: string | null): any | null {
  if (item === null) {
    return null;
  }

  try {
    return SuperJSON.parse(item);
  } catch (e) {
    console.warn("SuperJSON Parse error", e);
    return null;
  }
}

export interface PlayerFieldOverrides {
  name: string | null;
  loadout: LoadoutConfig | null;
  autoStart: boolean;
}

export interface DraggablePlayer {
  id: string;
  displayName: string;
  icon: string;
  info: BotInfo | PsyonixBotInfo | HumanInfo;
  tags: string[];
  uniquePathSegment?: string;
  overrides: PlayerFieldOverrides;
}

export interface ToggleableScript {
  id: string;
  displayName: string;
  icon: string;
  info: BotInfo;
  tags: string[];
  uniquePathSegment?: string;
}

export function draggablePlayerToPlayerJs(d: DraggablePlayer): PlayerJs {
  if (d.info instanceof BotInfo) {
    const player = BotInfo.createFrom(structuredClone(d.info));
    // Apply overrides
    player.config.settings.name =
      d.overrides.name ?? player.config.settings.name;
    player.loadout = d.overrides.loadout ?? d.info.loadout;
    if (!d.overrides.autoStart) {
      player.config.settings.runCommand = "";
      player.config.settings.runCommandLinux = "";
    }
    // We don't need to know the icon to start a bot.
    // This fixes oversized requests that result in a CORS error on windows (WebView2)
    player.icon = "";

    return {
      sort: "rlbot",
      player: player,
    };
  }

  if (d.info instanceof PsyonixBotInfo) {
    const player = PsyonixBotInfo.createFrom(structuredClone(d.info));
    // Apply overrides
    player.name = d.overrides.name ?? ""; // Empty names are replaced with random names for Psyonix bots
    player.loadout = d.overrides.loadout ?? d.info.loadout;

    return {
      sort: "psyonix",
      player: player,
    };
  }

  return {
    sort: "human",
    player: d.info,
  };
}

export default app;
