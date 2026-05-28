#!/usr/bin/env python3
"""对比 scripts/db.json 与 cubing.com WCA 列表的赛事 ID。

- db.json: WCA competition id，无横线（例如 AHAUOpen2019）。
- cubing API: https://cubing.com/api/competition 分页，`alias` 为带横线的 slug，
  WCA id 等价于去掉横线的 alias。

分页逻辑与 scripts/wca_province_stats.py 一致（year 必须为空）。
"""

from __future__ import annotations

import argparse
import json
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

BASE = "https://cubing.com/api/competition"
SCRIPT_DIR = Path(__file__).resolve().parent


def fetch_page(page: int, timeout: float) -> list[dict]:
    q = urllib.parse.urlencode(
        {
            "year": "",
            "type": "WCA",
            "province": "",
            "event": "",
            "page": str(page),
        }
    )
    url = f"{BASE}?{q}"
    req = urllib.request.Request(
        url,
        headers={
            "User-Agent": "cubingPro-db-cubing-diff/1.0",
            "Accept": "application/json",
        },
    )
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
    """分页拉取全部 WCA（与 wca_province_stats 相同）。"""
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
                f"[stderr] page={page} +{len(batch)} 累计={len(all_rows)}",
                file=sys.stderr,
            )
        page += 1
    return all_rows, pages_ok


def alias_to_wca_id(alias: str | None) -> str:
    if not alias:
        return ""
    return str(alias).replace("-", "")


def main() -> int:
    p = argparse.ArgumentParser(description="对比 db.json 与 cubing.com WCA 赛事 ID")
    p.add_argument(
        "--db",
        type=Path,
        default=SCRIPT_DIR / "db.json",
        help="本地 DBJSON 路径",
    )
    p.add_argument("--timeout", type=float, default=90.0)
    p.add_argument("-v", "--verbose", action="store_true")
    p.add_argument(
        "--json-out",
        type=Path,
        help="可选：写入完整对称差 JSON",
    )
    args = p.parse_args()

    try:
        with open(args.db, encoding="utf-8") as f:
            db_rows = json.load(f)
    except OSError as e:
        print(f"无法读取 {args.db}: {e}", file=sys.stderr)
        return 1
    except json.JSONDecodeError as e:
        print(f"{args.db} 不是合法 JSON: {e}", file=sys.stderr)
        return 1

    if not isinstance(db_rows, list):
        print("db.json 根应为数组", file=sys.stderr)
        return 1

    db_ids = {row["id"] for row in db_rows if isinstance(row, dict) and "id" in row}

    try:
        rows, pages_ok = fetch_all_competitions(args.timeout, args.verbose)
    except (urllib.error.URLError, urllib.error.HTTPError) as e:
        print(f"拉取 cubing API 失败: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"请求失败: {e}", file=sys.stderr)
        return 1

    cubing_by_canon: dict[str, str] = {}
    missing_alias = 0
    for row in rows:
        al = row.get("alias") if isinstance(row, dict) else None
        canon = alias_to_wca_id(al if isinstance(al, str) else None)
        if not canon:
            missing_alias += 1
            continue
        cubing_by_canon.setdefault(canon, al if isinstance(al, str) else "")

    cubing_canon = set(cubing_by_canon)
    only_in_db = sorted(db_ids - cubing_canon)
    only_in_cubing = sorted(cubing_canon - db_ids)

    print(f"db.json ID 数量: {len(db_ids)}")
    print(f"粗饼 type=WCA 条数(API 行): {len(rows)}，页数: {pages_ok}")
    print(f"粗饼有 slug（alias→去横线）的 WCA ID 数量: {len(cubing_canon)}（无 alias: {missing_alias}）")
    print(f"仅在 db.json: {len(only_in_db)}")
    print(f"仅在粗饼: {len(only_in_cubing)}")

    if args.json_out:
        out = {
            "db_path": str(args.db.resolve()),
            "db_id_count": len(db_ids),
            "cubing_rows": len(rows),
            "cubing_api_pages": pages_ok,
            "cubing_with_alias_count": len(cubing_canon),
            "cubing_missing_alias_count": missing_alias,
            "only_in_db_json": only_in_db,
            "only_in_cubing": only_in_cubing,
        }
        args.json_out.parent.mkdir(parents=True, exist_ok=True)
        args.json_out.write_text(json.dumps(out, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"\nJSON 已写入: {args.json_out.resolve()}")

    show = min(120, max(len(only_in_db), len(only_in_cubing)))
    if only_in_db:
        print(f"\n--- 仅在 db.json（前 min(120,{len(only_in_db)}) 条）---")
        for x in only_in_db[:120]:
            print(x)
    if only_in_cubing:
        print(f"\n--- 仅在粗饼（前 min(120,{len(only_in_cubing)}) 条）---")
        for x in only_in_cubing[:120]:
            print(x)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
