#!/usr/bin/env python3
"""从 cubing.com 拉取全部 WCA 赛事，统计各省（含直辖市）各地级市办赛情况。"""

from __future__ import annotations

import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from collections import Counter, defaultdict
from pathlib import Path

BASE = "https://cubing.com/api/competition"

DIRECT_MUNICIPALITIES = {"北京市", "天津市", "上海市", "重庆市"}

DATA_FILE = Path(__file__).parent / "china_province.data.txt"

CITY_ALIASES = {
    "鄂尔多斯": "伊克昭",
    "襄阳": "襄樊",
    "普洱": "思茅",
    "达州": "达川",
}


def fetch_page(page: int, timeout: float) -> list[dict]:
    q = urllib.parse.urlencode({
        "year": "", "type": "WCA", "province": "", "event": "", "page": str(page),
    })
    url = f"{BASE}?{q}"
    req = urllib.request.Request(url, headers={
        "User-Agent": "cubingPro-wca-city-stats/1.0",
        "Accept": "application/json",
    })
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        raw = resp.read().decode("utf-8")
    payload = json.loads(raw)
    if payload.get("status") != 0:
        raise RuntimeError(f"API status != 0: {payload!r}")
    data = payload.get("data")
    if not isinstance(data, list):
        raise RuntimeError(f"unexpected data type: {type(data)}")
    return data


def fetch_all_competitions(timeout: float, verbose: bool) -> tuple[list[dict], int]:
    all_rows: list[dict] = []
    prev_fp: tuple[int, int, int] | None = None
    page = 1
    pages_ok = 0
    while True:
        batch = fetch_page(page, timeout)
        if not batch:
            break
        fp = (batch[0]["id"], batch[-1]["id"], len(batch))
        if fp == prev_fp:
            if verbose:
                print(f"[stderr] 检测到重复分页，在 page={page} 停止", file=sys.stderr)
            break
        prev_fp = fp
        all_rows.extend(batch)
        pages_ok += 1
        if verbose:
            print(
                f"[stderr] page={page} 本页 {len(batch)} 条，累计 {len(all_rows)} 条",
                file=sys.stderr,
            )
        page += 1
    return all_rows, pages_ok


def load_province_cities(data_file: Path) -> dict[str, set[str]]:
    prov_cities: dict[str, set[str]] = {}
    with open(data_file, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("省"):
                continue
            parts = line.split("\t")
            if len(parts) < 3:
                continue
            prov, city, district = [p.strip() for p in parts[:3]]
            if not prov:
                continue
            if prov not in prov_cities:
                prov_cities[prov] = set()
            if city == "省直辖行政单位":
                prov_cities[prov].add(district)
            else:
                prov_cities[prov].add(city)
    return prov_cities


def normalize_prov(name: str) -> str:
    for suffix in ("自治区", "壮族自治区", "维吾尔自治区", "回族自治区", "藏族自治区"):
        if name.endswith(suffix):
            return name[:-len(suffix)]
    for suffix in ("省", "市"):
        if name.endswith(suffix):
            return name[:-len(suffix)]
    return name


def normalize_city(name: str) -> str:
    name = name.strip()
    name = re.sub(r"[（(].*?[）)]", "", name).strip()
    name = re.sub(r"\s+", "", name)
    name = re.sub(r"[?？]", "", name)
    if name.endswith("自治州"):
        return name[:-3]
    if name.endswith("地区"):
        return name[:-2]
    if name.endswith("盟"):
        return name[:-1]
    if name.endswith("市"):
        return name[:-1]
    return name


def main() -> int:
    p = argparse.ArgumentParser(description="统计各省WCA比赛各地级市办赛情况")
    p.add_argument("--timeout", type=float, default=30.0)
    p.add_argument("-v", "--verbose", action="store_true")
    p.add_argument("--json", action="store_true")
    args = p.parse_args()

    prov_cities = load_province_cities(DATA_FILE)

    city_lookup: dict[str, dict[str, str]] = {}
    for prov, cities in prov_cities.items():
        norm_prov = normalize_prov(prov)
        city_lookup[norm_prov] = {}
        for c in cities:
            nc = normalize_city(c)
            if nc and nc not in city_lookup[norm_prov]:
                city_lookup[norm_prov][nc] = c

    try:
        all_rows, pages_ok = fetch_all_competitions(args.timeout, args.verbose)
    except Exception as e:
        print(f"请求失败: {e}", file=sys.stderr)
        return 1

    city_counts: dict[str, Counter[str]] = defaultdict(Counter)
    unmatched: set[tuple[str, str]] = set()

    for comp in all_rows:
        for loc in comp.get("locations") or []:
            api_prov = (loc.get("province") or "").strip()
            api_city = (loc.get("city") or "").strip()
            if not api_prov or not api_city:
                continue

            lookup = city_lookup.get(api_prov)
            if not lookup:
                unmatched.add((api_prov, api_city))
                continue

            norm_city = normalize_city(api_city)
            aliased = CITY_ALIASES.get(norm_city, norm_city)
            full_city = lookup.get(norm_city) or lookup.get(aliased)
            if full_city:
                city_counts[api_prov][full_city] += 1
            else:
                unmatched.add((api_prov, api_city))

    if unmatched:
        print("[警告] 以下赛点未能匹配到数据文件中的城市:", file=sys.stderr)
        for prov, city in sorted(unmatched):
            print(f"  {prov} - {city}", file=sys.stderr)

    if args.json:
        result = {}
        for prov in sorted(prov_cities.keys()):
            cities = sorted(prov_cities[prov])
            norm_prov = normalize_prov(prov)
            counts = city_counts.get(norm_prov, Counter())
            held = {c: counts[c] for c in cities if c in counts}
            not_held = sorted(c for c in cities if c not in counts)
            result[prov] = {
                "held": dict(sorted(held.items(), key=lambda x: -x[1])),
                "not_held": not_held,
            }
        print(json.dumps(result, ensure_ascii=False, indent=2))
        return 0

    print(f"WCA 赛事总数: {len(all_rows)}（{pages_ok} 页）\n")
    for prov in sorted(prov_cities.keys()):
        cities = sorted(prov_cities[prov])
        norm_prov = normalize_prov(prov)
        counts = city_counts.get(norm_prov, Counter())
        is_direct = prov in DIRECT_MUNICIPALITIES

        held = sorted(
            [(c, counts[c]) for c in cities if c in counts], key=lambda x: -x[1]
        )
        not_held = [c for c in cities if c not in counts]

        held_str = ", ".join(f"{c}({n})" for c, n in held) if held else "-"
        not_held_str = ", ".join(not_held) if not_held else "-"

        if is_direct:
            print(f"{prov} | {held_str} | -")
        else:
            print(f"{prov} | {held_str} | {not_held_str}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
