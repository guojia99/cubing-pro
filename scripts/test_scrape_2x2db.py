import json
from pathlib import Path
import unittest
from collections import Counter

from scrape_2x2db import EXPECTED, PUZZLES, decode_string, displayed, extract_entries


class ScraperTests(unittest.TestCase):
    def test_string_decoding(self):
        self.assertEqual(decode_string(r'`R U\x27 \u0052 \` \\`'), "R U' R ` \\")

    def test_extracts_records_without_merging(self):
        bundle = ('var data=[{name:`Case (Alg 1)`,notation:`R U`,caseNotation:`R`,'
                  'set:`CLL`,subset:`S`,puzzle:`2x2`},'
                  '{name:`Case (Alg 2)`,notation:`R U`,caseNotation:`R`,'
                  'set:`CLL`,subset:`S`,puzzle:`2x2`}];')
        entries = extract_entries(bundle)
        self.assertEqual(len(entries), 2)
        self.assertEqual(entries[1]["name"], "Case (Alg 2)")

    def test_site_filter(self):
        for notation, want in (("B2", False), ("(U", False), ("y' (U", False),
                               ("R2 F2 R'", True), ("(U) R U", False),
                               ("(R U)3", True)):
            with self.subTest(notation=notation):
                self.assertEqual(displayed({"set": "TEG-2", "puzzle": "2x2", "notation": notation}), want)
        self.assertTrue(displayed({"set": "CLL", "puzzle": "2x2", "notation": "R"}))

    def test_every_published_record_matches_raw_source(self):
        root = Path(__file__).resolve().parents[1] / "build" / "Algs"
        raw = json.loads((root / "2x2db" / "source_entries.json").read_bytes())
        excluded = json.loads((root / "2x2db" / "excluded_entries.json").read_bytes())
        metadata = json.loads((root / "2x2db" / "source.json").read_bytes())
        self.assertEqual(len(raw), 10003)
        self.assertEqual(excluded, [entry for entry in raw if not displayed(entry)])
        self.assertEqual(len(excluded), 13)
        selected = [entry for entry in raw if displayed(entry)]
        self.assertEqual(dict(Counter(entry["puzzle"] for entry in selected)), EXPECTED)
        self.assertEqual(metadata["counts"], EXPECTED)
        catalog = json.loads((root / "catalog.json").read_bytes())
        imported = 0
        for item in catalog:
            if not item["path"].startswith("2x2db/"):
                continue
            _, directory, _ = item["path"].split("/")
            puzzle = next(key for key, value in PUZZLES.items() if value == directory)
            path = root / item["path"]
            actual = json.loads((path / "algs_info.json").read_bytes())
            set_name = next(iter(json.loads((path / "algsets_info.json").read_bytes())))
            expected = [entry for entry in selected if entry["puzzle"] == puzzle and entry["set"] == set_name]
            with self.subTest(set=item["name"]):
                self.assertEqual(len(actual), len(expected))
                self.assertEqual(len(actual), item["cases"])
                for index, entry in enumerate(expected, 1):
                    self.assertEqual(actual[str(index)], {
                        "a": [entry["notation"]], "name": entry["name"],
                        "group": entry["subset"], "algset": entry["set"],
                        "scramble": entry.get("caseNotation", entry["notation"]),
                    })
            imported += len(actual)
        self.assertEqual(imported, 9990)


if __name__ == "__main__":
    unittest.main()
