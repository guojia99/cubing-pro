"""Fetch 2x2DB with its display filter, preserving every raw source record."""

import argparse
from collections import Counter
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import re
import urllib.request


SITE = "https://2x2db.vercel.app"
PUZZLES = {"2x2": "2x2", "2x2 CBL": "2x2CBL", "EGLS": "EGLS", "Fake Bar": "FakeBar"}
EXPECTED = {"2x2": 4275, "2x2 CBL": 2183, "EGLS": 2226, "Fake Bar": 1306}
STRING = r'`(?:\\[\s\S]|[^`\\])*`|"(?:\\.|[^"\\])*"'
FIELD = re.compile(r'([A-Za-z_$][\w$]*):(' + STRING + r')')
OBJECT = re.compile(r'\{(?:[A-Za-z_$][\w$]*:(?:' + STRING + r'),?)+\}')


def decode_string(value):
    if value.startswith('"'):
        return json.loads(value)
    value = value[1:-1]
    if "${" in value:
        raise ValueError("Unexpected template interpolation in source data")
    escapes = {"n": "\n", "r": "\r", "t": "\t", "b": "\b", "f": "\f", "v": "\v", "0": "\0"}

    def unescape(match):
        token = match.group(1)
        if token.startswith(("u", "x")):
            return chr(int(token[1:], 16))
        if token in ("\n", "\r\n"):
            return ""
        return escapes.get(token, token)

    return re.sub(r'\\(u[0-9a-fA-F]{4}|x[0-9a-fA-F]{2}|\r?\n|[\s\S])', unescape, value)


def extract_entries(bundle):
    entries = []
    for start in re.finditer(r'(?:var |,)([A-Za-z_$][\w$]*)=\[(?=\{name:)', bundle):
        cursor = start.end()
        while True:
            match = OBJECT.match(bundle, cursor)
            if not match:
                raise ValueError(f"Unexpected data object at offset {cursor}")
            entry = {field.group(1): decode_string(field.group(2)) for field in FIELD.finditer(match.group())}
            if entry.get("puzzle") in PUZZLES:
                if not all(entry.get(key) for key in ("name", "notation", "set", "subset")):
                    raise ValueError(f"Incomplete source entry: {entry}")
                entries.append(entry)
            cursor = match.end()
            if bundle[cursor] == "]":
                break
            if bundle[cursor] != ",":
                raise ValueError(f"Unexpected array separator at offset {cursor}")
            cursor += 1
    if not entries:
        raise ValueError("No 2x2DB source entries found")
    return entries


def fetch(url):
    request = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0 (cubing-pro 2x2DB importer)"})
    with urllib.request.urlopen(request, timeout=40) as response:
        return response.read()


def displayed(entry):
    if entry["set"] != "TEG-2" and entry["puzzle"] != "Fake Bar":
        return True
    # Match the site's L()/z() default ETM count, including its leading AUF rule.
    notation = re.sub(r'^\([^)]*\)(?!\s*\d)', " ", entry["notation"].strip())
    repeated = []

    def expand(match):
        repeated.extend(match.group(1).strip().split() * int(match.group(2)))
        return " "

    notation = re.sub(r'\(([^)]*)\)\s*(\d+)', expand, notation)
    moves = repeated + re.sub(r'[()]', " ", notation).split()
    return len(moves) > 2


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=Path(__file__).resolve().parents[1] / "test" / "2x2db")
    args = parser.parse_args()
    html = fetch(SITE + "/database").decode()
    match = re.search(r'/assets/algorithms-[A-Za-z0-9._-]+\.js', html)
    if not match:
        raise ValueError("Algorithm bundle not found on database page")
    url = SITE + match.group()
    bundle = fetch(url)
    raw_entries = extract_entries(bundle.decode())
    entries = [entry for entry in raw_entries if displayed(entry)]
    excluded = [entry for entry in raw_entries if not displayed(entry)]
    counts = dict(Counter(entry["puzzle"] for entry in entries))
    print(json.dumps({"bundle": url, "raw_counts": dict(Counter(e["puzzle"] for e in raw_entries)), "displayed_counts": counts}, ensure_ascii=False))
    if counts != EXPECTED:
        raise ValueError(f"Source counts changed: expected {EXPECTED}, got {counts}; inspect before importing")
    catalog = {}
    for puzzle, directory in PUZZLES.items():
        by_set = {}
        for entry in entries:
            if entry["puzzle"] == puzzle:
                by_set.setdefault(entry["set"], []).append(entry)
        sets = []
        for name, records in by_set.items():
            slug = re.sub(r"[^A-Za-z0-9._+-]", "_", name)
            target = args.output / "data" / directory / slug
            algorithms, groups, scrambles = {}, {}, {}
            for index, record in enumerate(records, 1):
                key = str(index)
                setup = record.get("caseNotation", record["notation"])
                algorithms[key] = {"a": [record["notation"]], "name": record["name"],
                                   "group": record["subset"], "algset": name, "scramble": setup}
                groups.setdefault(record["subset"], []).append(index)
                scrambles[key] = [setup] if setup else []
            for filename, content in (("algs_info.json", algorithms), ("algsets_info.json", {name: list(groups)}),
                                      ("groups_info.json", groups), ("scrambles.json", scrambles)):
                write_json(target / filename, content)
            sets.append({"slug": slug, "name": name, "cases": len(records)})
        catalog[directory] = {"name": puzzle, "sets": sets}
    write_json(args.output / "catalog.json", catalog)
    write_json(args.output / "source_entries.json", raw_entries)
    write_json(args.output / "excluded_entries.json", excluded)
    write_json(args.output / "source.json", {
        "url": SITE + "/database", "bundle_url": url,
        "bundle_sha256": hashlib.sha256(bundle).hexdigest(),
        "fetched_at": datetime.now(timezone.utc).isoformat(),
        "counts": counts, "total": len(entries),
        "raw_counts": dict(Counter(e["puzzle"] for e in raw_entries)),
        "raw_total": len(raw_entries), "excluded_total": len(excluded),
        "policy": "One record per website-displayed entry; same TEG-2/Fake Bar ETM>2 filter as the site. No merging, deduplication, or name rewriting. All raw and excluded entries retained separately.",
    })
    print(f"Saved {len(entries)} source entries in {sum(len(v['sets']) for v in catalog.values())} sets; no images generated")


if __name__ == "__main__":
    main()
