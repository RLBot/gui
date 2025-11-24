import type { DraggablePlayer } from ".";
import { HumanInfo, PsyonixBotInfo } from "../bindings/gui";
import controller from "./assets/controller.svg";

export const BASE_PLAYERS: DraggablePlayer[] = [
  {
    displayName: "Human",
    icon: controller,
    id: crypto.randomUUID(),
    player: new HumanInfo(),
    tags: ["human"],
    overrides: { name: "Human", loadout: null, autoStart: true },
  },
  {
    displayName: "Psyonix Beginner",
    icon: "",
    id: crypto.randomUUID(),
    player: new PsyonixBotInfo({
      name: "",
      skill: 0,
    }),
    tags: ["psyonix"],
    overrides: { name: "", loadout: null, autoStart: true },
  },
  {
    displayName: "Psyonix Rookie",
    icon: "",
    id: crypto.randomUUID(),
    player: new PsyonixBotInfo({
      name: "",
      skill: 1,
    }),
    tags: ["psyonix"],
    overrides: { name: "", loadout: null, autoStart: true },
  },
  {
    displayName: "Psyonix Pro",
    icon: "",
    id: crypto.randomUUID(),
    player: new PsyonixBotInfo({
      name: "",
      skill: 2,
    }),
    tags: ["psyonix"],
    overrides: { name: "", loadout: null, autoStart: true },
  },
  {
    displayName: "Psyonix Allstar",
    icon: "",
    id: crypto.randomUUID(),
    player: new PsyonixBotInfo({
      name: "",
      skill: 3,
    }),
    tags: ["psyonix"],
    overrides: { name: "", loadout: null, autoStart: true },
  },
];
