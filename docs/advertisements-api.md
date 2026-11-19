# 广告接口

以下路径相对于 API 根路径 `/v3/cube-api`。管理接口使用现有 JWT + 超级管理员权限组，普通管理员不可访问；公开接口匿名访问，沿用公开接口限流。

| 方法 | 路径 | 成功响应 |
| --- | --- | --- |
| GET | `/admin/system_result/advertisements` | 直接返回完整 Advertisement 数组 |
| PUT | `/admin/system_result/advertisements` | 完整数组替换；`{"code":"200","data":null,"msg":"ok"}` |
| POST | `/admin/system_result/advertisements/image` | 图片上传，返回下述 ok envelope |
| GET | `/public/advertisements` | 直接返回 enabled 为 true 的 Advertisement 数组 |

## Advertisement

```json
[
  {
    "key": "shop-1",
    "title": "广告标题",
    "description": "广告说明",
    "contact": "联系方式",
    "imageUrl": "/v3/cube-api/static/image/uuid",
    "targetUrl": "https://example.com/shop",
    "duration": 5,
    "enabled": true,
    "products": [{"imageUrl": "/product.png", "caption": "商品", "href": ""}],
    "showcase": [{"imageUrl": "/showcase.png", "caption": "展示"}]
  }
]
```

- 字段名称均为 lowerCamelCase，无省略字段；products/showcase 缺失或 null 时规范为 `[]`。
- key 区分大小写，匹配 `[a-zA-Z0-9_-]{1,64}`，不可重复，不可使用 `cuberwenjun`、`contact`。
- title 去除空白后不可为空，最多 200 个 Unicode 字符；description 最多 4000 字符，contact 最多 500 字符。
- duration 为秒数，允许小数，范围 1..120；每次最多 50 条，每条 products/showcase 各最多 20 项。
- 所有 imageUrl 必填；targetUrl、products.href 可为空。非空 URL 仅允许 `/` 开头的站内绝对路径或带主机名的 http(s) URL，拒绝协议相对路径 `//...`、反斜杠、控制字符及相应百分号编码，不允许 URL 用户凭据。
- 管理 GET 保留禁用项和保存顺序；公开 GET 仅过滤禁用项。KV 无记录返回 `[]`，损坏数据或数据库异常返回错误；PUT `[]` 清空，`null` 无效。
- 数据保存在现有系统 KV 的 `__advertisements__`，无需新表。PUT 请求体最多 8 MiB，完整替换不提供并发版本控制，最后成功写入者生效。
- 参数错误返回现有 422 错误结构，请求体超限返回 413；存储异常返回 500，绝不以空数组掩盖异常。

## 图片上传

请求 JSON（字段大小写与头像上传一致）：

```json
{"ImageName":"cover.png","Data":"不含 data: 前缀的标准 base64"}
```

```json
{"code":"200","data":{"uid":"uuid","url":"/v3/cube-api/static/image/uuid","width":1200,"height":600},"msg":"ok"}
```

- 只接受 JPEG、PNG 实际图像内容；解码后原文件不超过 5 MiB，单边不超过 8192 像素，总像素不超过 2000 万。解析 JSON 前限制请求体为 5 MiB 的 base64 编码长度加 64 KiB 开销（支持 chunked 请求）。
- 验证图像尺寸并完整解码成功后，原样保存 base64 解码得到的文件字节，不重新编码，保留 JPEG EXIF Orientation 和 PNG 色彩等元数据。文件路径为 `${api.staticPath}/advertisements/<uuid>.<jpeg|png>`。ImageName 不参与文件路径，也不作为可信格式来源。
- 响应 width/height 为文件编码的原始像素尺寸，不应用 EXIF 旋转；例如原始 1200×600、Orientation=6 的 JPEG 返回 width=1200、height=600，支持 EXIF 的客户端按方向信息显示。
- 复用 system.Image，Use 为 `advertisement`，记录上传者 ID；数据库写入失败清理新文件；不修改用户或头像。
- 通过现有匿名 `/static/image/:uid` 读取；公开图片接口移除错误的 `image/.png` 类型头，由标准文件服务确定 Content-Type。
- `api.staticPath` 是 API 图片根目录，独立于网关前端静态目录。空值与头像保持一致，使用进程工作目录；生产环境应由部署方决定持久化路径、写权限与备份策略，此功能不设置生产路径。
- 本接口没有图片删除/回收功能；替换广告配置不会删除已上传图片，旧图片 URL 仍可访问。禁用广告只影响公开列表。
