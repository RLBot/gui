import "./global.css";
import { mount } from "svelte";
import {
  BotInfo,
  HumanInfo, LoadoutConfig,
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

export interface PlayerFieldOverrides {
  name: string;
  loadout: LoadoutConfig | null
  autoStart: boolean;
}

export interface DraggablePlayer {
  id: string;
  displayName: string;
  icon: string;
  player: BotInfo | PsyonixBotInfo | HumanInfo;
  tags: string[];
  uniquePathSegment?: string;
  overrides: PlayerFieldOverrides;
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
  if (d.player instanceof BotInfo) {
    let player = BotInfo.createFrom(structuredClone(d.player))
    // Apply overrides
    player.config.settings.name = d.overrides.name;
    player.loadout = d.overrides.loadout ?? d.player.loadout;
    if (!d.overrides.autoStart) {
      player.config.settings.runCommand = ""
      player.config.settings.runCommandLinux = ""
    }
    // We don't need to know the icon to start a bot.
    // This fixes oversized requests that result in a CORS error on windows (WebView2)
    player.config.settings.logoFile = "";

    return {
      sort: "rlbot",
      player: player,
    }
  }

  if (d.player instanceof PsyonixBotInfo) {
    let player = PsyonixBotInfo.createFrom(structuredClone(d.player))
    // Apply overrides
    player.name = d.overrides.name;
    player.loadout = d.overrides.loadout ?? d.player.loadout;

    return {
      sort: "psyonix",
      player: player
    }
  }

  return {
    sort: "human",
    player: d.player,
  };
}

export default app;
