# SpeedDB 公式爬虫

支持浏览器访问 / HTTP 请求 / 本地 HTML 文件，解析 SpeedDB 公式页面。

## 安装

```bash
pip install -r requirements.txt
playwright install chromium
```

## 用法

```bash
# 使用浏览器访问在线 URL (默认，会等待内容完全加载)
python speeddb.py "https://www.speedcubedb.com/a/3x3/F2L"

# 延长等待时间 (默认 15 秒)
python speeddb.py "https://www.speedcubedb.com/a/3x3/F2L" --wait 30

# 使用 HTTP 请求 (不启动浏览器)
python speeddb.py "https://www.speedcubedb.com/a/3x3/F2L" --http

# 从本地 HTML 文件解析
python speeddb.py speeddb.html

# 强制仅从本地文件解析
python speeddb.py speeddb.html --file-only

# 不请求 More Algorithms 接口
python speeddb.py speeddb.html --no-more

# 输出到文件 (默认 output.json)
python speeddb.py "https://www.speedcubedb.com/a/3x3/F2L" -o result.json
```

## 输出格式

每个公式输出为：

```json
{
    "name": "F2L 1",
    "algs": {
        "Front Right": ["U R U' R'", "R' F R F'", ...],
        "Front Left": ["F' r U r'", "d R U' R'", ...],
        "Back Left": [...],
        "Back Right": [...]
    },
    "setup": "F R' F' R",
    "group": "Free Pairs",
    "image": "<svg xmlns=...>...</svg>"
}
```

- **name**: 公式名称 (data-alg)
- **algs**: 按分组 (Front Right / Front Left / Back Left / Back Right) 的公式列表，含 More Algorithms 拉取的更多公式
- **setup**: 初始 setup
- **group**: 子分组 (data-subgroup)
- **image**: SVG 图片的完整 HTML

## 自适应

分组名称从页面 `.tabs-orientation .subcatname` 解析，若页面结构不同会以第一条公式为准，适配不同公式类型。
