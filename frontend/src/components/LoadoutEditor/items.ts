//@ts-ignore
import ItemsCsv from "../../assets/items.csv";
import { ITEM_TYPES } from "./itemtypes";

export interface CsvItem {
  id: number;
  uuid: string;
  name: string;
}

function toTitleCase(str: string) {
  return str
    .replaceAll("_", " ")
    .replace(
      /\w\S*/g,
      (text) => text.charAt(0).toUpperCase() + text.substring(1).toLowerCase(),
    )
    .trim();
}

function renameItem(item: CsvItem, category: string) {
  const uuidParts = item.uuid.split(".").pop();
  if (!uuidParts) return item.name;

  let specifier = uuidParts.toLowerCase().replace(category.toLowerCase(), "");

  const nameParts = item.name.split(":");
  if (nameParts.length > 1) {
    const bodyName = nameParts[0].toLowerCase();
    specifier = specifier.replace(bodyName, "");
  }

  return `${item.name} (${toTitleCase(specifier)})`;
}

// this is "Black Market Drop", "Exotic Drop" etc that have the category "Body" for some reason
const EXCLUDE_ITEMS = [5364, 5365, 5366, 5367, 5368, 5369];

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
    const id = +columns[0];

    if (items[category] && !EXCLUDE_ITEMS.includes(id))
      items[category].push({
        id,
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

    // sort category items by name
    items[category].sort((a, b) => a.name.localeCompare(b.name));
  }

  return items;
}
