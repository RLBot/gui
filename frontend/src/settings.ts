import { Window } from "@wailsio/runtime";
import toast from "svelte-5-french-toast";
import { writable } from "svelte/store";
import { MAPS_STANDARD } from "./arena-names";

export const THEMES = {
  "Light blurred": ["lightTheme", "blur"],
  Light: ["lightTheme"],
  "Dark blurred": ["darkTheme", "blur"],
  Dark: ["darkTheme"],
  "Follow system": null,
} as const;

export const currentTheme = writable(
  (localStorage.getItem("APP_THEME") as keyof typeof THEMES) || "Follow system",
);
currentTheme.subscribe((value) => {
  console.log("Theme is now", value);
  localStorage.setItem("APP_THEME", value);
});

export const mapStore = writable(
  localStorage.getItem("MS_MAP") || MAPS_STANDARD["DFH Stadium"],
);
mapStore.subscribe((value) => {
  localStorage.setItem("MS_MAP", value);
});

export const showWindowControls = writable(
  (localStorage.getItem("FRAMELESS") ?? "true") !== "false",
);
let init = true;
showWindowControls.subscribe((value) => {
  if (value) {
    if (!init) {
      toast.success("You need to relaunch RLBot to fully apply frameless mode", {
        duration: 4000
      })
    }
  } else {
    Window.SetFrameless(value)
  }
  init = false;
  localStorage.setItem("FRAMELESS", value.toString());
});

export const zoomLevel = writable(
  Number(localStorage.getItem("ZOOM")) || 1,
);
zoomLevel.subscribe((value) => {
  Window.SetZoom(value)
});
