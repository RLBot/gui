import type { TeamLoadoutConfig, TeamPaintConfig } from "../../../bindings/gui";

export interface ItemType {
  name: string;
  category: string;
  itemKey: keyof TeamLoadoutConfig;
  paintKey: keyof TeamPaintConfig | null;
}

export const ITEM_TYPES: ItemType[] = [
  {
    name: "Body",
    category: "Body",
    itemKey: "carId",
    paintKey: "carPaintId",
  },
  {
    name: "Decal",
    category: "Skin",
    itemKey: "decalId",
    paintKey: "decalPaintId",
  },
  {
    name: "Wheels",
    category: "Wheels",
    itemKey: "wheelsId",
    paintKey: "wheelsPaintId",
  },
  {
    name: "Boost",
    category: "Boost",
    itemKey: "boostId",
    paintKey: "boostPaintId",
  },
  {
    name: "Antenna",
    category: "Antenna",
    itemKey: "antennaId",
    paintKey: "antennaPaintId",
  },
  {
    name: "Topper",
    category: "Hat",
    itemKey: "hatId",
    paintKey: "hatPaintId",
  },
  {
    name: "Primary Finish",
    category: "PaintFinish",
    itemKey: "paintFinishId",
    paintKey: null,
  },
  {
    name: "Accent Finish",
    category: "PaintFinish",
    itemKey: "customFinishId",
    paintKey: null,
  },
  {
    name: "Engine Audio",
    category: "EngineAudio",
    itemKey: "engineAudioId",
    paintKey: null,
  },
  {
    name: "Trail",
    category: "SupersonicTrail",
    itemKey: "trailsId",
    paintKey: "trailsPaintId",
  },
  {
    name: "Goal Explosion",
    category: "GoalExplosion",
    itemKey: "goalExplosionId",
    paintKey: "goalExplosionPaintId",
  },
];
