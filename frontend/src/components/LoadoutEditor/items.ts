//@ts-ignore
import ItemsCsv from "../../assets/items.csv";
import { ITEM_TYPES } from "./itemtypes";

export interface CsvItem {
  id: number;
  uuid: string;
  name: string;
}

function toTitleCase(str: string) {
  return str.replace(
    /\w\S*/g,
    (text) => text.charAt(0).toUpperCase() + text.substring(1).toLowerCase(),
  );
}

function renameItem(item: CsvItem, category: string) {
  let specifier = item.uuid.split(".").pop();
  if (!specifier) return item.name;

  specifier = specifier
    .toLowerCase()
    .replace(category.toLowerCase(), "")
    .replaceAll("_", " ");

  const nameParts = item.name.split(":");
  if (nameParts.length > 1) {
    const bodyName = nameParts[0].toLowerCase();
    specifier = specifier.replace(bodyName, "");
  }

  return `${item.name} (${toTitleCase(specifier.trim())})`;
}

export async function getAndParseItems() {
  const resp = await fetch(ItemsCsv);
  const csv = await resp.text();
  const lines = csv.split(/\r?\n/);

  const items: {
    [x: string]: CsvItem[];
  } = {};

  for (const key in ITEM_TYPES) {
    const category = ITEM_TYPES[key].category;
    items[category] = [];
  }

  for (const line of lines) {
    const columns = line.split(",");
    const category = columns[1];

    if (items[category])
      items[category].push({
        id: +columns[0],
        uuid: columns[2],
        name: columns[3],
      });
  }

  // rename duplicate item names
  for (const category in items) {
    const nameCounts: { [x: string]: [boolean, number] } = {};

    for (const [i, item] of items[category].entries()) {
      if (nameCounts[item.name]) {
        const [needsHandling, idx] = nameCounts[item.name];

        item.name = renameItem(item, category);

        // rename the original item as well
        if (needsHandling) {
          const item = items[category][idx];
          nameCounts[item.name] = [false, idx];
          item.name = renameItem(item, category);
        }
      } else {
        nameCounts[item.name] = [true, i];
      }
    }
  }

  return items;
}
