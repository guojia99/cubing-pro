#!/usr/bin/env python3
"""Render 26 weeks of repository activity using only Git and Python's stdlib."""

import argparse
from datetime import date, datetime, timedelta, timezone
from html import escape
import math
from pathlib import Path
import subprocess


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "docs/assets/activity.svg"
BOT_EMAIL = "41898282+github-actions[bot]@users.noreply.github.com"
WEEKS = 26


def git(*args):
    return subprocess.check_output(["git", *args], cwd=ROOT, text=True).strip()


def collect(as_of):
    if git("rev-parse", "--is-shallow-repository") == "true":
        raise SystemExit("Full Git history required; run git fetch --unshallow first.")
    monday = as_of - timedelta(days=as_of.weekday())
    start = monday - timedelta(weeks=WEEKS - 1)
    counts = [0] * WEEKS
    authors = set()
    for line in git("log", "HEAD", "--no-merges", "--format=%ct%x09%aE").splitlines():
        timestamp, email = line.split("\t", 1)
        day = datetime.fromtimestamp(int(timestamp), timezone.utc).date()
        if start <= day <= as_of and email.casefold() != BOT_EMAIL.casefold():
            counts[(day - start).days // 7] += 1
            authors.add(email.casefold())
    return start, counts, len(authors)


def render(as_of, start, counts, authors):
    left, right, top, bottom = 60, 1060, 184, 320
    ceiling = max(4, math.ceil(max(counts) / 4) * 4)
    points = [
        (left + i * (right - left) / (WEEKS - 1), bottom - n / ceiling * (bottom - top))
        for i, n in enumerate(counts)
    ]
    line = "M" + " L".join(f"{x:.1f},{y:.1f}" for x, y in points)
    area = f"{line} L{right},{bottom} L{left},{bottom} Z"
    description = (
        f"{sum(counts)} non-merge commits by {authors} authors across "
        f"{sum(n > 0 for n in counts)} active weeks, {start} through {as_of} UTC. "
        "README chart bot excluded. Current week may be incomplete."
    )
    parts = [f'''<svg xmlns="http://www.w3.org/2000/svg" width="1120" height="396" viewBox="0 0 1120 396" role="img" aria-labelledby="title desc">
  <title id="title">Cubing Pro · Repository activity</title>
  <desc id="desc">{escape(description)}</desc>
  <defs>
    <linearGradient id="area" x1="0" y1="0" x2="0" y2="1">
      <stop stop-color="#22d3ee" stop-opacity=".28" />
      <stop offset="1" stop-color="#22d3ee" stop-opacity=".02" />
    </linearGradient>
  </defs>
  <style>
    .bg {{ fill: #0d192b; }}
    .text {{ fill: #f1f5f9; }}
    .muted {{ fill: #94a3b8; }}
    .grid {{ stroke: #94a3b8; stroke-opacity: .17; }}
    .line {{ stroke: #22d3ee; }}
    .point {{ fill: #22d3ee; }}
    text {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; }}
    @media (prefers-color-scheme: light) {{
      .bg {{ fill: #f4f9fc; }}
      .text {{ fill: #0f172a; }}
      .muted {{ fill: #475569; }}
      .line {{ stroke: #0891b2; }}
      .point {{ fill: #0891b2; }}
    }}
  </style>
  <rect class="bg" x=".5" y=".5" width="1119" height="395" rx="18" stroke="#94a3b8" stroke-opacity=".2" />
  <text class="text" x="40" y="42" font-size="18" font-weight="700">Repository activity</text>
  <text class="muted" x="1080" y="42" font-size="12" text-anchor="end">26 WEEKS / WEEKLY COMMITS</text>''']
    for x, value, label in [
        (40, sum(counts), "COMMITS"),
        (340, authors, "AUTHORS"),
        (640, sum(n > 0 for n in counts), "ACTIVE WEEKS"),
    ]:
        parts.append(f'''  <text class="text" x="{x}" y="104" font-size="36" font-weight="700">{value}</text>
  <text class="muted" x="{x}" y="129" font-size="11" letter-spacing="1.4">{label}</text>''')
    for i in range(5):
        y = bottom - i * (bottom - top) / 4
        parts.append(f'''  <path class="grid" d="M{left},{y:.1f} H{right}" />
  <text class="muted" x="44" y="{y + 4:.1f}" text-anchor="end" font-size="10">{ceiling * i // 4}</text>''')
    parts.append(f'  <path d="{area}" fill="url(#area)" />')
    parts.append(f'  <path class="line" d="{line}" fill="none" stroke-width="2.5" stroke-linejoin="round" />')
    for i, ((x, y), count) in enumerate(zip(points, counts)):
        week = start + timedelta(weeks=i)
        parts.append(f'  <circle class="point" cx="{x:.1f}" cy="{y:.1f}" r="3"><title>Week of {week}: {count} commits</title></circle>')
        if i % 5 == 0:
            parts.append(f'  <text class="muted" x="{x:.1f}" y="344" text-anchor="middle" font-size="11">{week:%b %d}</text>')
    parts.append(f'''  <text class="muted" x="40" y="378" font-size="11">Non-merge commits · chart bot excluded · UTC · current week incomplete</text>
  <text class="muted" x="1080" y="378" font-size="11" text-anchor="end">Updated {as_of}</text>
</svg>
''')
    return "\n".join(parts)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--as-of", type=date.fromisoformat, default=datetime.now(timezone.utc).date())
    args = parser.parse_args()
    start, counts, authors = collect(args.as_of)
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT.write_text(render(args.as_of, start, counts, authors), encoding="utf-8")
    print(f"Updated {OUTPUT.relative_to(ROOT)}: {sum(counts)} commits, {authors} authors")


if __name__ == "__main__":
    main()
