#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.9"
# ///
"""Convert the repository item list from Windows-1252 to UTF-8."""

from pathlib import Path


ITEMS_CSV = Path(__file__).resolve().parents[1] / "frontend/src/assets/items.csv"


def main() -> None:
    source = ITEMS_CSV.read_bytes()
    converted = source.decode("cp1252").encode("utf-8")
    ITEMS_CSV.write_bytes(converted)
    print(f"Re-encoded {ITEMS_CSV} as UTF-8 ({len(converted)} bytes)")


if __name__ == "__main__":
    main()
