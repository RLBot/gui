import { uuidv4, type DraggablePlayer } from ".";
import { HumanInfo, PsyonixBotInfo } from "../bindings/gui";
import controller from "./assets/controller.svg";

export const BASE_PLAYERS: DraggablePlayer[] = [
  {
    displayName: "Human",
    icon: controller,
    id: uuidv4(),
    player: new HumanInfo(),
    tags: ["human"],
  },
  {
    displayName: "Psyonix Beginner",
    icon: "",
    id: uuidv4(),
    player: new PsyonixBotInfo({
      skill: 0,
    }),
    tags: ["psyonix"],
  },
  {
    displayName: "Psyonix Rookie",
    icon: "",
    id: uuidv4(),
    player: new PsyonixBotInfo({
      skill: 1,
    }),
    tags: ["psyonix"],
  },
  {
    displayName: "Psyonix Pro",
    icon: "",
    id: uuidv4(),
    player: new PsyonixBotInfo({
      skill: 2,
    }),
    tags: ["psyonix"],
  },
  {
    displayName: "Psyonix Allstar",
    icon: "",
    id: uuidv4(),
    player: new PsyonixBotInfo({
      skill: 3,
    }),
    tags: ["psyonix"],
  },
];
