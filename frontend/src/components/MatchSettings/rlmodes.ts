import { mutators } from "./rlmutators";

export type Gamemode = {
  match: { [x: string]: string };
  mutators: { [x: string]: string };
};

export const gamemodes: {
  [x: string]: Gamemode;
} = {
  "Heatseeker Ricochet": {
    match: {
      game_mode: mutators.game_mode[5],
      game_map_upk: "Labs_PillarGlass_P",
    },
    mutators: {},
  },
  "Gotham City Rumble": {
    match: {
      game_mode: mutators.game_mode[0],
      game_map_upk: "Park_Bman_P",
    },
    mutators: {
      rumble: mutators.rumble[10],
    },
  },
  "Speed Demon": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      boost_amount: mutators.boost_amount[1],
      boost_strength: mutators.boost_strength[2],
      ball_max_speed: mutators.ball_max_speed[3],
      ball_bounciness: mutators.ball_bounciness[1],
      ball_size: mutators.ball_size[3],
      respawn_time: mutators.respawn_time[2],
      demolish: mutators.demolish[3],
    },
  },
  "Ghost Hunt": {
    match: {
      game_mode: mutators.game_mode[0],
      game_map_upk: "Haunted_TrainStation_P",
    },
    mutators: {
      ball_type: mutators.ball_type[6],
      rumble: mutators.rumble[8],
      game_event: mutators.game_event[1],
      audio: mutators.audio[1],
    },
  },
  "Dropshot Rumble": {
    match: {
      game_mode: mutators.game_mode[2],
      game_map_upk: "ShatterShot_P",
    },
    mutators: {
      rumble: mutators.rumble[1],
    },
  },
  "Nike Fc Showdown": {
    match: {
      game_mode: mutators.game_mode[0],
      game_map_upk: "swoosh_p",
    },
    mutators: {
      ball_max_speed: mutators.ball_max_speed[2],
      ball_weight: mutators.ball_weight[6],
      ball_bounciness: mutators.ball_bounciness[4],
      ball_type: mutators.ball_type[7],
    },
  },
  "Tactical Rumble": {
    match: {
      game_mode: mutators.game_mode[4],
    },
    mutators: {
      rumble: mutators.rumble[9],
    },
  },
  "Gforce Frenzy": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      boost_amount: mutators.boost_amount[1],
      boost_strength: mutators.boost_strength[3],
      gravity: mutators.gravity[1],
    },
  },
  "Spike Rush": {
    match: {
      game_mode: mutators.game_mode[0],
      game_map_upk: "ThrowbackStadium_P",
    },
    mutators: {
      rumble: mutators.rumble[7],
      respawn_time: mutators.respawn_time[2],
      game_event: mutators.game_event[2],
    },
  },
  "Spring Loaded": {
    match: {},
    mutators: {
      rumble: mutators.rumble[5],
    },
  },
  "Spooky Cube": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      ball_max_speed: mutators.ball_max_speed[3],
      ball_type: mutators.ball_type[8],
      ball_weight: mutators.ball_weight[1],
      ball_bounciness: mutators.ball_bounciness[2],
      boost_amount: mutators.boost_amount[1],
    },
  },
  "Beach Ball": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      ball_max_speed: mutators.ball_max_speed[2],
      ball_type: mutators.ball_type[4],
      ball_weight: mutators.ball_weight[5],
      ball_size: mutators.ball_size[2],
      ball_bounciness: mutators.ball_bounciness[2],
    },
  },
  "Super Cube": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      ball_max_speed: mutators.ball_max_speed[3],
      ball_type: mutators.ball_type[1],
      ball_weight: mutators.ball_weight[1],
      ball_bounciness: mutators.ball_bounciness[2],
      boost_amount: mutators.boost_amount[1],
    },
  },
  "Boomer Ball": {
    match: {
      game_mode: mutators.game_mode[0],
    },
    mutators: {
      boost_amount: mutators.boost_amount[1],
      boost_strength: mutators.boost_strength[1],
      ball_max_speed: mutators.ball_max_speed[3],
      ball_bounciness: mutators.ball_bounciness[2],
      ball_weight: mutators.ball_weight[3],
    },
  },
};
