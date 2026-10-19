import astra from "./assets/rockethost/Astra.webp";
import chronos from "./assets/rockethost/ChronosV2.1.webp";
import daizen from "./assets/rockethost/Daizen.webp";
import disco from "./assets/rockethost/Disco-3.webp";
import ducky from "./assets/rockethost/Ducky.webp";
import fury from "./assets/rockethost/Fury.webp";
import galactus from "./assets/rockethost/GalactusV2.webp";
import kiyo from "./assets/rockethost/Kiyo3.1.webp";
import necto from "./assets/rockethost/Necto.webp";
import nexto from "./assets/rockethost/Nexto.webp";
import opti from "./assets/rockethost/Opti.webp";
import optiFR from "./assets/rockethost/Opti-FR.webp";
import optiGP from "./assets/rockethost/Opti-GP-V2.2.webp";
import ordis from "./assets/rockethost/Ordis.webp";
import ripple from "./assets/rockethost/Ripple.webp";
import wisp from "./assets/rockethost/Wispv2.webp";
import zeus from "./assets/rockethost/ZeusV3.webp";
import defaultLogo from "./assets/rlbot_mono.png";

// Use canonical server names. Versions in each group share one bundled logo.
const logoGroups: { names: string[]; logo: string }[] = [
  { names: ["AstraV5"], logo: astra },
  { names: ["Chronos"], logo: chronos },
  { names: ["Daizen"], logo: daizen },
  { names: ["Discombobulated-2-3", "Discombobulated-3-4"], logo: disco },
  { names: ["Ducky"], logo: ducky },
  { names: ["Fury"], logo: fury },
  { names: ["Galactus"], logo: galactus },
  { names: ["Kiyo", "Kiyo3"], logo: kiyo },
  { names: ["necto"], logo: necto },
  { names: ["nexto", "nextoplus"], logo: nexto },
  { names: ["Opti"], logo: opti },
  { names: ["OptiFR"], logo: optiFR },
  { names: ["OptiGP"], logo: optiGP },
  { names: ["Ordis"], logo: ordis },
  { names: ["RippleV1", "RippleV1.1"], logo: ripple },
  { names: ["Wisp"], logo: wisp },
  { names: ["Zeus"], logo: zeus },
];

const logos = new Map(
  logoGroups.flatMap(({ names, logo }) =>
    names.map((name) => [name, logo] as const),
  ),
);

export function rocketHostLogo(name: string): string {
  return logos.get(name) ?? defaultLogo;
}

export function isDefaultRocketHostLogo(name: string): boolean {
  return !logos.has(name);
}
