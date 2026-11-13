"""Import the selected trainer data from a pinned upstream ZIP archive."""

import argparse
import json
from pathlib import Path
from zipfile import ZipFile


REVISION = "3d19a613932319e33219105ad5bfd1f7a128ff24"
TRAINERS = {
    **{f"FTO-{name}": f"FTO-{name}" for name in (
        "LL", "LT", "LBT", "L3C", "BTLT", "FTLT", "L6X"
    )},
    "FTO-L3T": "Octaminx-L3T",
    "FTO-TCP": "Octaminx-TCP",
    "5x5-LCE": "5x5-LCE",
}
FILES = (
    "algs_info.json", "algsets_info.json", "groups_info.json",
    "combined.json", "scrambles.json",
)


def merge_pairs(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            previous = result[key]
            if not isinstance(previous, dict) or not isinstance(value, dict):
                raise ValueError(f"Unexpected duplicate key: {key}")
            for field, field_value in value.items():
                if field in previous and previous[field] != field_value:
                    raise ValueError(f"Conflicting duplicate field: {key}.{field}")
                previous[field] = field_value
        else:
            result[key] = value
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("archive", type=Path)
    parser.add_argument("--check", action="store_true", help="Verify existing files without writing")
    args = parser.parse_args()
    destination = Path(__file__).resolve().parents[1] / "build" / "Alg-Trainers"
    prefix = f"Alg-Trainers-{REVISION}/"
    pending = {}
    with ZipFile(args.archive) as archive:
        for source, target in TRAINERS.items():
            for name in FILES:
                data = archive.read(f"{prefix}{source}-Trainer/{name}")
                if source == "FTO-LT" and name == "algs_info.json":
                    merged = json.loads(data, object_pairs_hook=merge_pairs)
                    if len(merged) != 151 or any(not case.get("a") for case in merged.values()):
                        raise ValueError("FTO-LT must contain 151 cases with algorithms")
                    data = (json.dumps(merged, ensure_ascii=False, indent=2) + "\n").encode()
                else:
                    json.loads(data)
                pending[destination / f"{target}-Trainer" / name] = data
        pending[destination / "LICENSE"] = archive.read(prefix + "LICENSE")

    for path, data in pending.items():
        if args.check:
            if path.read_bytes() != data:
                raise ValueError(f"Local file differs from the pinned import: {path}")
        else:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
    action = "Verified" if args.check else "Imported"
    print(f"{action} {len(TRAINERS)} trainers ({len(pending)} files) from {REVISION}")


if __name__ == "__main__":
    main()
