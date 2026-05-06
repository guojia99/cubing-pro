#!/usr/bin/env python3
"""
从 cubing.com 拉取 WCA 赛事列表，统计指定年份各省（含直辖市）比赛场次。

重要：该接口在 URL 中带上非空 year 时，分页会失效（每页重复相同数据）。
因此脚本始终使用 year= 空串分页，再按「赛历开始日」在本地时区下的年份过滤。

口径：每个赛事的 `locations` 里，每个出现一次的省份计 1 场（多地联办多省各计 1）。
参赛人数取接口字段 `registered_competitors`；全国总参赛人数按场次去重（每场只加一次），
分省列按赛点重复计入多地联赛人数（与场数口径一致）。

API 形态与前端一致：
  https://cubing.com/api/competition?year=&type=WCA&province=&event=&page=1
"""

from __future__ import annotations

import argparse
import json
import sys
import urllib.error
import urllib.parse
import urllib.request
from collections import Counter
from datetime import datetime

try:
    from zoneinfo import ZoneInfo
except ImportError:  # Python < 3.9 理论上少见
    ZoneInfo = None  # type: ignore[misc, assignment]


BASE = "https://cubing.com/api/competition"


def fetch_page(page: int, timeout: float) -> list[dict]:
    """拉取一页（year 必须为空，分页才有效）。"""
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
            "User-Agent": "cubingPro-wca-province-stats/1.0",
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
    """分页拉取全部 WCA 赛事，遇到空页或与上一页完全相同的指纹时停止。

    返回 (赛事列表, 实际成功读取的页数)。
    """
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


def start_year_local(comp: dict, tz) -> int:
    d = comp.get("date") or {}
    fr = int(d.get("from", 0))
    return datetime.fromtimestamp(fr, tz=tz).year


def stats_by_province(
    comps: list[dict], expect_year: int, tz
) -> tuple[Counter[str], Counter[str]]:
    """按省份统计场数与参赛人数（registered_competitors）。

    场次口径与文档一致；参赛人数按赛点省份分摊：同一赛事在多个省份出现时，
    各省累加该场的 registered_competitors（合计行单独给出全国每场仅计一次的总和）。
    """
    events: Counter[str] = Counter()
    registered: Counter[str] = Counter()
    for comp in comps:
        if start_year_local(comp, tz) != expect_year:
            continue
        reg = int(comp.get("registered_competitors") or 0)
        locs = comp.get("locations") or []
        if not locs:
            events["（无赛点信息）"] += 1
            registered["（无赛点信息）"] += reg
            continue
        seen: set[str] = set()
        for loc in locs:
            prov = (loc.get("province") or "").strip()
            if not prov:
                prov = "（未填省份）"
            if prov in seen:
                continue
            seen.add(prov)
            events[prov] += 1
            registered[prov] += reg
    return events, registered


def total_registered_in_year(comps: list[dict], expect_year: int, tz) -> int:
    """该年每场赛事的 registered_competitors 之和（每场只计一次）。"""
    total = 0
    for comp in comps:
        if start_year_local(comp, tz) != expect_year:
            continue
        total += int(comp.get("registered_competitors") or 0)
    return total


def main() -> int:
    p = argparse.ArgumentParser(description="统计 cubing.com 上某年各省 WCA 比赛场次")
    p.add_argument(
        "--year",
        type=int,
        default=datetime.now().year,
        help="年份，默认当年（按赛历开始日、见 --timezone）",
    )
    p.add_argument(
        "--timezone",
        default="Asia/Shanghai",
        help="判断「今年」用的 IANA 时区，默认中国常用",
    )
    p.add_argument("--timeout", type=float, default=30.0, help="单次 HTTP 超时（秒）")
    p.add_argument(
        "--json",
        action="store_true",
        help="输出 JSON（省份 -> 场数），否则打印表格",
    )
    p.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="分页进度输出到 stderr",
    )
    args = p.parse_args()
    year = args.year

    if ZoneInfo is None:
        print("需要 Python 3.9+（zoneinfo）", file=sys.stderr)
        return 1
    tz = ZoneInfo(args.timezone)

    try:
        all_rows, pages_ok = fetch_all_competitions(args.timeout, args.verbose)
    except urllib.error.HTTPError as e:
        print(f"HTTP 错误: {e}", file=sys.stderr)
        return 1
    except urllib.error.URLError as e:
        print(f"网络错误: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"请求失败: {e}", file=sys.stderr)
        return 1

    counts, reg_by_province = stats_by_province(all_rows, year, tz)
    in_year = sum(1 for comp in all_rows if start_year_local(comp, tz) == year)
    total_site_rows = sum(counts.values())
    total_registered = total_registered_in_year(all_rows, year, tz)

    if args.json:
        out = {
            "year": year,
            "timezone": args.timezone,
            "api_pages_scanned": pages_ok,
            "competitions_in_year": in_year,
            "total_registered_competitors": total_registered,
            "province_counts": dict(counts.most_common()),
            "province_registered_competitors": {
                prov: reg_by_province[prov] for prov, _ in counts.most_common()
            },
        }
        print(json.dumps(out, ensure_ascii=False, indent=2))
        return 0

    print(f"年份（按 {args.timezone} 赛历开始日）: {year}")
    print(f"分页拉取赛事总数: {len(all_rows)}（其中 {in_year} 场开始日在该年）")
    print(f"按赛点省份计场次数合计: {total_site_rows}")
    print(f"总参赛人数（registered_competitors，每场计一次）: {total_registered}")
    print()
    print(f"{'省份':<12} {'场数':>6} {'参赛人数':>10}")
    print("-" * 30)
    for prov, n in counts.most_common():
        r = reg_by_province[prov]
        print(f"{prov:<12} {n:>6} {r:>10}")
    print("-" * 30)
    print(f"{'合计':<12} {total_site_rows:>6} {total_registered:>10}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
