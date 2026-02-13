# -*- coding: utf-8 -*-
"""
SpeedDB 公式页面爬虫
支持: 浏览器访问 / HTTP 请求 / 本地 HTML 文件
"""

import json
import re
import sys
import time
from pathlib import Path
from typing import Optional
from urllib.parse import quote

import requests
from bs4 import BeautifulSoup
from tqdm import tqdm


def _log(msg: str, level: str = "INFO") -> None:
    """输出日志到 stderr，避免干扰 JSON 输出"""
    icons = {"INFO": "✓", "STEP": "→", "WARN": "⚠"}
    icon = icons.get(level, " ")
    print(f"{icon} {msg}", file=sys.stderr)

# 默认分组名称 (data-ori 0,1,2,3 对应)
DEFAULT_GROUP_NAMES = ["Front Right", "Front Left", "Back Left", "Back Right"]

BASE_URL = "https://www.speedcubedb.com"
CATEGORY_API = "https://www.speedcubedb.com/category.algs.php"


def get_category_url(alg_name: str, tab_id: int, cat: str) -> str:
    """
    More Algorithms 按钮请求的 category 接口 URL
    alg_name: data-algname, 如 "F2L 1"
    tab_id: data-d, 0-3 对应 Front Right, Front Left, Back Left, Back Right
    cat: data-category, 如 "F2L"
    """
    params = {
        "algname": alg_name,
        "d": tab_id,
        "cat": cat,
    }
    query = "&".join(f"{k}={quote(str(v))}" for k, v in params.items())
    return f"{CATEGORY_API}?{query}"


def fetch_more_algorithms(alg_name: str, tab_id: int, cat: str, max_retries: int = 5) -> list[str]:
    """请求 category 接口获取更多公式，返回公式字符串列表，失败时重试最多 5 次"""
    url = get_category_url(alg_name, tab_id, cat)
    last_error = None
    for attempt in range(max_retries):
        try:
            resp = requests.get(url, timeout=15)
            resp.raise_for_status()
            soup = BeautifulSoup(resp.text, "html.parser")
            algs = []
            for li in soup.select("li.list-group-item"):
                alg_el = li.select_one(".formatted-alg")
                if alg_el:
                    text = alg_el.get_text(strip=True)
                    if text:
                        algs.append(text)
            return algs
        except Exception as e:
            last_error = e
            if attempt < max_retries - 1:
                time.sleep(1)  # 重试前等待 1 秒
    _log(f"获取更多公式失败 (已重试 {max_retries} 次) {url}: {last_error}", "WARN")
    return []


def parse_setup(setup_el) -> str:
    """从 setup-case 元素解析 setup 文本"""
    if not setup_el:
        return ""
    text = setup_el.get_text(separator=" ", strip=True)
    # 移除 "setup:" 前缀
    if text.lower().startswith("setup:"):
        text = text[6:].strip()
    return text


def parse_image_from_block_html(block_html: str) -> str:
    """从 block 的原始 HTML 字符串中直接提取 svg 元素"""
    m = re.search(r"<svg[\s\S]*?</svg>", block_html)
    return m.group(0) if m else ""


def get_group_names_from_tabs(algo_block) -> list[str]:
    """
    从第一个公式的 tabs-orientation 获取分组名称，实现自适应。
    若未找到则使用默认分组名。
    """
    tabs = algo_block.select(".tabs-orientation .subcatname")
    if tabs:
        return [t.get_text(strip=True) for t in tabs if t.get_text(strip=True)]
    return DEFAULT_GROUP_NAMES.copy()


def parse_algorithms_from_page(
    soup: BeautifulSoup,
    raw_html: str = "",
    fetch_more: bool = True,
    category: Optional[str] = None,
    verbose: bool = True,
) -> list[dict]:
    """
    解析页面中所有公式块，返回结构化数据列表。
    支持页面内多个 singlealgorithm 块（如 F2L 1~41），按 data-alg 正确识别并去重。
    fetch_more: 是否请求 More Algorithms 接口获取更多公式
    category: 用于 More Algorithms 的 data-category，若为 None 则从页面解析
    verbose: 是否输出进度和日志
    """
    # 优先使用带 data-alg 的块，确保是公式块；若无则回退到 .singlealgorithm
    blocks = soup.select("div.singlealgorithm[data-alg]")
    if not blocks:
        blocks = soup.select(".singlealgorithm")
    if not blocks:
        if verbose:
            _log("未发现公式块", "WARN")
        return []

    # 按 data-alg 去重，保留首次出现（保持 DOM 顺序），记录原始索引
    seen_alg: set[str] = set()
    unique_blocks: list = []
    unique_indices: list[int] = []
    for i, b in enumerate(blocks):
        alg = b.get("data-alg") or ""
        if alg and alg in seen_alg:
            continue
        if alg:
            seen_alg.add(alg)
        unique_blocks.append(b)
        unique_indices.append(i)

    if verbose:
        _log(f"发现 {len(unique_blocks)} 个公式块，开始解析", "STEP")

    # 从原始 HTML 提取每个 block 的 raw 内容，用于正确提取 svg
    block_raw_list: list[str] = []
    if raw_html:
        starts = [m.start() for m in re.finditer(
            r'<div[^>]*singlealgorithm[^>]*data-alg="[^"]*"[^>]*>',
            raw_html,
        )]
        block_raw_list = [
            raw_html[starts[i]: starts[i + 1]] if i + 1 < len(starts) else raw_html[starts[i]:]
            for i in range(len(starts))
        ]

    results = []
    group_names_template = None  # 以第一条公式的分组为准

    iterator = tqdm(unique_blocks, desc="解析公式", unit="个", file=sys.stderr, disable=not verbose)
    for list_idx, block in enumerate(iterator):
        raw_idx = unique_indices[list_idx] if list_idx < len(unique_indices) else list_idx
        name = block.get("data-alg") or ""
        group = block.get("data-subgroup") or ""
        if verbose and name:
            iterator.set_postfix_str(name)

        # Setup
        setup_el = block.select_one(".setup-case")
        setup = parse_setup(setup_el)

        # Image：直接从 block 原始 HTML 提取 svg
        block_html = block_raw_list[raw_idx] if raw_idx < len(block_raw_list) else str(block)
        image = parse_image_from_block_html(block_html)

        # 获取分组名称 (首次从当前块解析，后续用模板)
        if group_names_template is None:
            group_names_template = get_group_names_from_tabs(block)
        group_names = group_names_template

        # 获取 category (用于 More Algorithms)
        cat = category
        if not cat:
            link = block.select_one("a[data-category]")
            if link:
                cat = link.get("data-category") or ""

        algs: dict[str, list[str]] = {gn: [] for gn in group_names}

        # 解析每个 data-ori 下的公式 (仅 div，排除 nav 中的 a 链接)
        ori_divs = block.select("div[data-ori]")
        for ori_div in ori_divs:
            ori_val = ori_div.get("data-ori")
            if ori_val is None:
                continue
            try:
                ori_idx = int(ori_val)
            except ValueError:
                continue
            if ori_idx >= len(group_names):
                continue
            group_name = group_names[ori_idx]

            # 当前 tab 下的公式 (排除 More Algorithms 所在的 li)
            for li in ori_div.select("ul.list-group > li.list-group-item"):
                alg_el = li.select_one(".formatted-alg")
                if alg_el:
                    text = alg_el.get_text(strip=True)
                    if text:
                        algs[group_name].append(text)

            # 请求 More Algorithms 获取更多公式
            if fetch_more and cat and name:
                more_btn = ori_div.select_one("button.more-algs")
                if more_btn:
                    more_algs = fetch_more_algorithms(name, ori_idx, cat)
                    algs[group_name].extend(more_algs)

        # 去重并保持顺序
        for k in algs:
            seen = set()
            unique = []
            for a in algs[k]:
                if a not in seen:
                    seen.add(a)
                    unique.append(a)
            algs[k] = unique

        results.append({
            "name": name,
            "algs": algs,
            "setup": setup,
            "group": group,
            "image": image,
        })

    return results


def crawl_with_browser(
    url_or_path: str,
    fetch_more: bool = True,
    verbose: bool = True,
    headless: bool = False,
    wait_seconds: int = 15,
) -> list[dict]:
    """
    使用浏览器访问页面，等待内容完全加载后爬取公式。
    wait_seconds: 页面加载后额外等待秒数，确保懒加载等内容加载完成
    """
    from playwright.sync_api import sync_playwright

    path = Path(url_or_path)
    if path.exists() and path.suffix.lower() in (".html", ".htm"):
        target = path.resolve().as_uri()
    else:
        target = url_or_path

    if verbose:
        _log("步骤 1/3: 使用浏览器打开页面...", "STEP")

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=headless)
        try:
            page = browser.new_page()
            page.goto(target, wait_until="networkidle", timeout=120000)
            if verbose:
                _log("页面已加载，等待内容完全渲染...", "INFO")

            # 多次滚动到底部，触发懒加载
            for _ in range(3):
                page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
                page.wait_for_timeout(2000)
            page.evaluate("window.scrollTo(0, 0)")
            page.wait_for_timeout(1000)

            # 等待指定时间，确保所有内容加载完成
            if wait_seconds > 0:
                if verbose:
                    _log(f"额外等待 {wait_seconds} 秒...", "INFO")
                page.wait_for_timeout(wait_seconds * 1000)

            # 等待公式块出现
            try:
                page.wait_for_selector(".singlealgorithm", timeout=10000)
            except Exception:
                pass

            html = page.content()
        finally:
            browser.close()

    if verbose:
        _log(f"页面获取完成 ({len(html):,} 字节)", "INFO")
        _log("步骤 2/3: 正在解析页面...", "STEP")

    soup = BeautifulSoup(html, "html.parser")
    return parse_algorithms_from_page(soup, raw_html=html, fetch_more=fetch_more, verbose=verbose)


def crawl_via_http(url: str, fetch_more: bool = True, verbose: bool = True) -> list[dict]:
    """
    通过 HTTP API 直接请求页面 HTML 并爬取公式。
    url: 网页 URL，如 https://www.speedcubedb.com/a/3x3/F2L
    """
    if verbose:
        _log("步骤 1/3: 正在请求页面...", "STEP")
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.9",
    }
    resp = requests.get(url, headers=headers, timeout=30)
    resp.raise_for_status()
    resp.encoding = resp.apparent_encoding or "utf-8"
    if verbose:
        _log(f"页面获取完成 ({len(resp.text):,} 字节)", "INFO")
        _log("步骤 2/3: 正在解析页面...", "STEP")
    soup = BeautifulSoup(resp.text, "html.parser")
    return parse_algorithms_from_page(soup, raw_html=resp.text, fetch_more=fetch_more, verbose=verbose)


def crawl_from_html_file(
    html_path: str,
    fetch_more: bool = True,
    verbose: bool = True,
) -> list[dict]:
    """
    直接从本地 HTML 文件解析 (不启动浏览器)。
    适用于已有完整 HTML 的离线解析。
    """
    path = Path(html_path)
    if not path.exists():
        raise FileNotFoundError(f"文件不存在: {html_path}")
    if verbose:
        _log("步骤 1/3: 正在读取本地文件...", "STEP")
    html = path.read_text(encoding="utf-8", errors="replace")
    if verbose:
        _log(f"文件读取完成 ({len(html):,} 字节)", "INFO")
        _log("步骤 2/3: 正在解析页面...", "STEP")
    soup = BeautifulSoup(html, "html.parser")
    return parse_algorithms_from_page(soup, raw_html=html, fetch_more=fetch_more, verbose=verbose)


def main():
    import argparse

    parser = argparse.ArgumentParser(description="SpeedDB 公式爬虫")
    parser.add_argument(
        "source",
        nargs="?",
        default="speeddb.html",
        help="URL 或本地 HTML 文件路径",
    )
    parser.add_argument(
        "--no-more",
        action="store_true",
        help="不请求 More Algorithms 接口",
    )
    parser.add_argument(
        "--file-only",
        action="store_true",
        help="仅从本地文件解析，不发起请求",
    )
    parser.add_argument(
        "--http",
        action="store_true",
        help="使用 HTTP 请求而非浏览器（默认用浏览器）",
    )
    parser.add_argument(
        "--headless",
        action="store_true",
        help="浏览器无头模式",
    )
    parser.add_argument(
        "--wait",
        type=int,
        default=15,
        metavar="SECONDS",
        help="浏览器加载后额外等待秒数 (默认: 15)",
    )
    parser.add_argument(
        "-o", "--output",
        default="output.json",
        help="输出 JSON 文件路径 (默认: output.json)",
    )
    parser.add_argument(
        "-q", "--quiet",
        action="store_true",
        help="静默模式，不输出进度和日志",
    )

    args = parser.parse_args()
    verbose = not args.quiet

    source = args.source
    is_url = source.startswith("http://") or source.startswith("https://")
    path = Path(source)
    is_local_html = path.exists() and path.suffix.lower() in (".html", ".htm")

    if args.file_only or (is_local_html and not is_url):
        results = crawl_from_html_file(source, fetch_more=not args.no_more, verbose=verbose)
    elif args.http:
        results = crawl_via_http(source, fetch_more=not args.no_more, verbose=verbose)
    else:
        results = crawl_with_browser(
            source,
            fetch_more=not args.no_more,
            verbose=verbose,
            headless=args.headless,
            wait_seconds=args.wait,
        )

    if verbose:
        _log("步骤 3/3: 解析完成", "STEP")
        _log(f"共解析 {len(results)} 个公式", "INFO")
        print(file=sys.stderr)  # 与 JSON 输出分隔

    out = json.dumps(results, ensure_ascii=False, indent=2)
    output_path = args.output
    Path(output_path).write_text(out, encoding="utf-8")
    if verbose:
        _log(f"结果已保存到: {output_path}", "INFO")
    print(out)


if __name__ == "__main__":
    main()
