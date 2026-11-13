"""Import the local test snapshots without their image payloads."""

import argparse
import hashlib
import json
from pathlib import Path


FILES = ("algs_info.json", "algsets_info.json", "groups_info.json", "scrambles.json")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    destination = root / "build" / "Algs"
    for source in ("2x2db", "algs"):
        if not (root / "test" / source / "catalog.json").is_file():
            parser.error(f"Original snapshot test/{source} is required for import/source comparison; it is not tracked in Git. Runtime loading and Go tests use build/Algs only.")
    entries = []
    pending = {}

    def add(source, puzzle, slug, cube, name, replaces):
        relative = Path(source) / puzzle / slug
        source_dir = root / "test" / source / "data" / puzzle / slug
        cases = json.loads((source_dir / "algs_info.json").read_bytes())
        hashes = {}
        for filename in FILES:
            data = (source_dir / filename).read_bytes()
            json.loads(data)
            pending[destination / relative / filename] = data
            hashes[filename] = hashlib.sha256(data).hexdigest()
        entries.append({"cube": cube, "name": name, "path": relative.as_posix(),
                        "replaces": replaces, "cases": len(cases),
                        "source": f"test/{source}/data/{puzzle}/{slug}", "sha256": hashes})

    catalog = json.loads((root / "test" / "2x2db" / "catalog.json").read_bytes())
    for puzzle, category in catalog.items():
        for item in category["sets"]:
            slug = item["slug"]
            name = {"LEG-1": "LEG", "EG1LS": "EGLS1", "EG2LS": "EGLS2",
                    "Fake_Bar": "FakeBar"}.get(slug, slug.replace("-", "") if slug != "TCLL-" else slug)
            replaces = {"CLL": ["EG"], "FH": ["FH"], "LS-1": ["LS"],
                        "TCLL+": ["TCLL"], "TEG-1": ["TEG"]}.get(slug, [])
            add("2x2db", puzzle, slug, "222", name, replaces)

    catalog = json.loads((root / "test" / "algs" / "catalog.json").read_bytes())
    for puzzle in ("3x3", "4x4", "5x5", "6x6", "pyraminx", "sq1"):
        for item in catalog[puzzle]:
            slug = item["slug"]
            name = item["en"]
            replaces = []
            if puzzle == "3x3":
                cube = "333oh" if slug == "oh-cmll" else "333"
                name = "CMLL" if slug == "oh-cmll" else name
                if slug in ("f2l", "oll", "pll", "coll", "cmll", "zbll", "zbls", "oh-cmll"):
                    replaces = [name]
            elif puzzle in ("4x4", "5x5", "6x6"):
                cube = "bigCube"
                suffix = {"oll-parity": "OLLP", "pll-parity": "PLLP", "l2e": "L2E", "l2c": "L2C"}[slug]
                name = f"{puzzle} {suffix}"
                replaces = {("4x4", "pll-parity"): ["PLLP"], ("5x5", "l2e"): ["L2E"]}.get((puzzle, slug), [])
            elif puzzle == "pyraminx":
                if slug != "l4e":
                    continue
                cube, name, replaces = "special", "Pyraminx L4E", ["Pyraminx L4E"]
            else:
                cube, name = "sq1", slug.upper()
                replaces = [name] if slug in ("cs", "co", "eo", "obl", "pbl") else []
                if slug in ("cp", "ep"):
                    replaces = ["CPEP"]
            add("algs", puzzle, slug, cube, name, replaces)

    pending[destination / "catalog.json"] = (json.dumps(entries, ensure_ascii=False, indent=2) + "\n").encode()
    for name in ("source.json", "source_entries.json", "excluded_entries.json"):
        source = root / "test" / "2x2db" / name
        pending[destination / "2x2db" / name] = source.read_bytes()
    for path, data in pending.items():
        if args.check:
            if path.read_bytes() != data:
                raise ValueError(f"Imported file differs from source: {path}")
        else:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
    print(f"{'Verified' if args.check else 'Imported'} {len(entries)} sets, {sum(e['cases'] for e in entries)} cases, {len(pending)} files")


if __name__ == "__main__":
    main()
