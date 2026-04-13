# cubing-pro 后端 `src` 目录说明

本文档描述 Go 后端源码根目录 `src/` 的层级结构及各文件/目录职责。根路径前缀均为 `src/`。

**约定**：`api/app/` 下多数文件为「一个文件 ≈ 一个 HTTP 处理函数（Gin Handler）」，文件名与导出函数名通常对应业务动作（如 `CreateComp.go` → 创建比赛）。

---

## 顶层目录树（不含第三方子树细节）

```
src/
├── api/                 # HTTP API（Gin）：路由、中间件、业务 Handler、公开封装
├── configs/             # YAML 配置结构体与加载
├── gateway/             # 网关：HTTPS/静态资源/tNoodle 代理等
├── internel/            # 核心业务：DB 模型、打乱、算法库、定时任务、工具等（包名拼写为 internel）
├── robot/               # QQ / CQHTTP 机器人入口与插件
├── staticx/             # 与静态统计/实验性 DB 相关的辅助包
├── test_tool/           # 开发/测试辅助（如内存监控）
└── wca/                 # 本地 WCA 数据访问、同步、城市静态数据等
```

以下按目录展开说明。

---

## `configs/`

| 文件 | 作用 |
|------|------|
| `configs.go` | 定义 `GlobalConfig`、`APIConfig`、`GatewayConfig`、`QQBotConfig`、`WcaAuth2` 等配置结构，以及从 YAML 加载配置的 `Load` 方法。 |

---

## `gateway/`

| 文件 | 作用 |
|------|------|
| `gateway.go` | 网关服务：Gin 实例、静态文件扩展名策略、反向代理 API/tNoodle、安全头（secure）、gzip 等；与前端部署、HTTPS 证书路径等相关。 |

---

## `api/`

### `api/api.go`

| 文件 | 作用 |
|------|------|
| `api.go` | 组装 Gin：`/v3/cube-api` 路由组，挂载全局中间件（日志、Recovery、CORS），初始化 JWT/鉴权，注册各 `routes` 子模块。 |

### `api/routes/`

将业务按 URL 分组注册到 Gin；各函数名即路由分组职责。

| 文件 | 作用 |
|------|------|
| `public.go` | 公开接口：赛事、玩家、统计、通知、论坛、组织等无需登录或限流可访问的路由。 |
| `auth.go` | 登录、注册、验证码、WCA OAuth、用户信息更新等认证相关路由。 |
| `admin.go` | 管理端路由（用户管理、系统配置等，具体见文件内注册）。 |
| `user.go` | 普通登录用户相关路由。 |
| `comp_result.go` | 赛事与成绩相关（组织者/用户侧赛事操作，与 `public` 中只读列表互补）。 |
| `post.go` | 帖子、板块、主题等社区路由。 |
| `sports.go` | 非 WCA「运动/扩展项目」成绩与事件路由。 |
| `wca.go` | WCA 查询类 API（选手、国家、中国赛事等）；含 `GET /wca/cubing-china/person/:wcaID`（粗饼选手页代理：IP 限流 + Handler 内全局锁与最小出站间隔）。 |
| `static.go` | 静态资源或上传图片等路由。 |

### `api/middleware/`

| 文件 | 作用 |
|------|------|
| `cors.go` | 跨域中间件。 |
| `jwt.go` | JWT 解析与注入上下文。 |
| `check_auth.go` | 登录态/权限校验。 |
| `check_header.go` | 请求头校验（若启用）。 |
| `cache.go` | 响应缓存中间件。 |
| `rate_limit.go` | 限流。 |
| `log.go` | 请求日志辅助。 |
| `code.go` | 与业务状态码/响应码相关的中间件或工具。 |
| `utils.go` | 中间件共用工具函数。 |

### `api/exception/`

| 文件 | 作用 |
|------|------|
| `base.go` | 异常/错误基类或统一结构。 |
| `code.go` | 错误码定义。 |
| `ok.go` | 成功响应封装。 |

### `api/public/`

对「公开 API」逻辑的再封装，供 `routes` 引用。

| 文件 | 作用 |
|------|------|
| `alg.go` | 公开算法库/Trainer 相关接口。 |
| `events.go` | 公开项目/赛事事件列表等。 |
| `user.go` | 公开用户信息片段。 |
| `wca.go` | 公开 WCA 查询封装。 |

### `api/utils/`

| 文件 | 作用 |
|------|------|
| `bind.go` | 请求体绑定、参数校验辅助。 |
| `list.go` | 分页、列表参数处理等。 |

### `api/app/`（按子包）

| 子目录 | 职责概要 |
|--------|----------|
| `acknowledgments/` | 赞助/鸣谢列表的读、写（`Get*` / `Set*` / `types`）。 |
| `auth/` | 注册、登录、验证码、找回/重置密码、头像、用户消息、资料更新、`wca.go`（WCA OAuth）、`WCA_AUTH_README.md`（说明文档）。 |
| `comp/` | 用户侧：比赛详情、报名、退赛、成绩列表、报名进度与回调等。 |
| `events/` | 赛事内「项目（Events）」的创建、删除、列表（组织者侧逻辑，与路由绑定一致）。 |
| `notify/` | 站内通知的增删改查与列表。 |
| `organizers/` | 主办方：创建/更新/删除比赛、选手与成绩录入、预审、结束比赛、组织成员等；`org_mid/middleware.go` 为组织者相关中间件。 |
| `other_link/` | 站点外链配置的读写与类型定义。 |
| `pktimer/` | PK 计时相关 HTTP 接口。 |
| `post/` | 论坛：板块、主题、帖子、封禁主题等。 |
| `result/` | 玩家成绩、预录入成绩、导出、宿敌、SOR、时间报表等。 |
| `sports/` | 扩展运动项目：事件与成绩的增删查。 |
| `statics/` | 静态图片等资源接口（如 `Image.go`）。 |
| `statistics/` | 榜单与统计：`best`、`records`、`kinch`、`sor`、`diy_rankings` 等。 |
| `systemResult/` | 站点级配置：标题、Logo、页脚、欢迎语、键值等读写。 |
| `users/` | 用户列表、详情、封禁、管理员重置密码、键值等。 |
| `wca/` | WCA 相关业务接口：中国赛事、`player`、`country`、`statics`、`cubing_china_person.go`（粗饼选手主页 JSON）等。 |

各子目录内 `.go` 文件一般为单一 Handler，见上文「约定」。

---

## `internel/`

> 包名 `internel` 为项目历史拼写，全项目统一引用此路径。

### `internel/svc/`

| 文件 | 作用 |
|------|------|
| `svc.go` | 服务容器：`Svc` 聚合 `*gorm.DB`、配置、`Convenient`、打乱组件、`wca.WCA`、内存缓存等；`NewAPISvc` 负责读配置、连库、初始化打乱与算法 Trainer、WCA 与定时任务门面。 |

### `internel/database/model/`

GORM 模型定义，按业务域分子包；`base/base_model.go` 为公共字段基类。

| 子目录 | 作用 |
|--------|------|
| `algdb/` | 算法库配置模型。 |
| `competition/` | 比赛、报名、分组、讨论等。 |
| `crawler/` | 爬虫侧发邮件等模型。 |
| `event/` | 项目定义、路线、对外 DTO（`out.go`）。 |
| `pktimer/` | PK 计时相关表。 |
| `post/` | 帖子、通知表单。 |
| `result/` | 成绩、预成绩、记录、`results_utils` 与测试。 |
| `sports/` | 运动扩展成绩。 |
| `system/` | 系统键值、图片等。 |
| `user/` | 用户、角色、OAuth state、校验、赞助关联等。 |
| `wca/` | 与 WCA 同步或 API 结果相关的表；`utils/results.go` 为结果工具与测试。 |
| `README.md` | 模型层说明（若存在）。 |

### `internel/convenient/`

| 文件/目录 | 作用 |
|-----------|------|
| `convenient.go` | 「便捷门面」实现，聚合业务用例供 `Svc` 使用。 |
| `interface/` | 接口定义：`i_user`、`i_competition`、`i_result`、`i_wca_results` 及结果类型/工具抽象，便于测试与解耦。 |
| `job/` | 定时任务：更新 Cubing China、DIY 排行、记录任务等（`job.go` 为调度入口或汇总）。 |

### `internel/scramble/`

| 文件 | 作用 |
|------|------|
| `scramble.go` | 打乱服务主逻辑与接口实现。 |
| `tnoodle_scramble.go` | 对接 tNoodle 的打乱实现。 |
| `auto_scramble.go` | 自动/备用打乱路径。 |
| `scramble_images.go` | 打乱图生成。 |
| `rust_scramble.go` / `rust_scramble_darwin.go` | 通过 CGO/静态库调用 Rust 打乱（平台相关）。 |
| `rust_scramble.h` / `librust_scramble.a` | Rust 静态库头文件与二进制（构建产物，版本库可能包含）。 |
| `*_test.go` | 对应包的单元测试。 |

### `internel/algdb/`

魔方公式数据库（多项目：222/333、Pyraminx、BLD、SQ1 CSP 等）。

| 文件模式 | 作用 |
|----------|------|
| `cube_*_db.go` | 各阶/项目公式数据访问。 |
| `alg_db.go` / `cube_db.go` | 通用或聚合入口。 |
| `bigbld_db.go` / `bld_db.go` | 盲拧相关。 |
| `sq1_csp_db.go` | SQ1 CSP。 |
| `cube_py_db.go` | Pyraminx 等。 |
| `mihlefeld_type.go` | 类型或枚举辅助。 |
| `script/` | `commutator.go` 等交换子/脚本辅助；含 `.js`/`.py` 为工具脚本。 |

### `internel/algs/`

Trainer 算法集、表驱动与读取逻辑。

| 文件 | 作用 |
|------|------|
| `algs.go` | 初始化与对外 API。 |
| `consts.go` / `types.go` / `trainers_types.go` / `tables.go` | 常量、类型、Trainer 元数据与表。 |
| `trainer_reader.go` | 从磁盘读取 Trainer 数据。 |
| `speeddb/` | SpeedDB 生成的 JSON/HTML/Python 工具及 `to_tr.go` 转换脚本；各 `*-Trainer` 子目录为算法集数据文件。 |

### `internel/wca_api/`

与 WCA 公开 API 或 seniors 等扩展数据交互。

| 文件 | 作用 |
|------|------|
| `person.go` / `person_types.go` / `person_result.go` | 人员与成绩 DTO 及请求。 |
| `db.go` | 本地缓存或 DB 访问辅助。 |
| `wcaSeniors*.go` | Seniors 相关类型与数据；`*_test.go` 为测试；`wca_seniors_cache.json` 等为缓存数据。 |
| `cubing_citys.go` | 城市相关。 |

### `internel/crawler/`

| 目录 | 作用 |
|------|------|
| `cubing/` | 爬取 Cubing 赛事/城市等：`cubing2_competition`、`wca_competition*`、`person_page.go`（粗饼选手主页抓取与解析，无锁；串行+节流在 API）、`competition_urls.json`、`cubing_city/` 下城市解析与测试数据。 |
| `sora/` | 其他爬虫扩展（若为空或占位以实际文件为准）。 |

### `internel/email/`

| 文件 | 作用 |
|------|------|
| `email.go` | 邮件发送封装。 |
| `EmailMsg.go` | 邮件内容与模板组装。 |
| `parser_file.html` / `base_email_msg.gohtml` | HTML 模板。 |
| `EmailMsg_test.go` | 测试。 |

### `internel/utils/`

通用工具：ID、时间、加解密、随机、字符串、JSON、HTTP、图片、切片、指针等；各 `*_test.go` 为单元测试。

### `internel/ttf/`

| 文件 | 作用 |
|------|------|
| `ttf.go` | 字体加载或使用（如成绩图、图片渲染）。 |
| `HuaWenHeiTi.ttf` | 华文黑体字体文件。 |

---

## `wca/`

本地 WCA 数据层（MySQL/SQLite 等）、同步与静态榜单数据。

| 文件/目录 | 作用 |
|-----------|------|
| `wca.go` | `WCA` 接口与实现入口，封装查询、统计、导出 SQLite 等。 |
| `sync.go` / `sync_*.go` | 与 WCA 数据库同步、静态数据导入 DB 等。 |
| `select.go` / `static.go` / `tools.go` | 查询辅助、静态聚合、工具函数；配套 `*_test.go`。 |
| `consts.go` | 常量。 |
| `types/` | `wca_types.go`、`static_types.go` 等领域类型。 |
| `utils/` | `bitset`、`top_heap`、结果工具等。 |
| `citys_data/*.json` | 中/省/市/区等行政区划静态 JSON。 |
| `test.json` 等 | 测试或样例数据。 |

---

## `staticx/`

| 文件 | 作用 |
|------|------|
| `staticx.go` / `types.go` | 静态统计实验或辅助结构（含本地 MySQL 连接示例，多用于开发/统计脚本）。 |
| `statistics_cubing_china_2024.go` | 特定年度中国区统计逻辑与测试。 |
| `static_cubing_event_dnf.go` | 与 DNF 等事件相关的静态统计。 |

---

## `robot/`

| 文件/目录 | 作用 |
|-----------|------|
| `client.go` | 根据配置创建 CQHTTP / QQ 官方机器人实例并 `Run`，挂载插件与 pktimer 客户端。 |
| `README.md` | 机器人模块说明。 |
| `types/` | 机器人消息、插件接口类型定义。 |
| `robots/` | **核心业务**：`robots.go` 消息路由与插件调度；`go_bot.go`、`cqhttp.go`、`cqhttp_types.go` 为各协议适配；`plugins.go` 注册插件；`plugin/p_*.go` 为具体指令（比赛、绑定、成绩、排行等）；`pktimer/` 与 QQ 侧 PK 计时联动；`tools/t_*.go` 为打乱、WCA、随机等工具命令。 |
| `qq_bot/Better-Bot-Go/` | 第三方 QQ 频道/机器人 Go SDK（dto、openapi、websocket 等），**一般作为依赖使用，不必逐文件修改**。 |
| `qq_bot/Bot-Client-Go/safe_ws/` | WebSocket 安全连接与日志初始化（QQ 客户端相关）。 |

---

## `test_tool/`

| 文件 | 作用 |
|------|------|
| `mem.go` | `MemMonitor`：测试时周期性打印内存使用，辅助排查泄漏或峰值。 |

---

## 与其他仓库路径的关系

- 可执行入口在仓库根目录 `cubing-pro/main.go`，通过 `cmd/` 子命令启动 `api`、`robot`、`gateway` 等；**不在 `src/` 内**，但与 `src` 紧密配合。
- 配置文件示例在 `cubing-pro/etc/`、`local/` 等，由 `configs` 加载。

---

## 维护说明

- 新增 HTTP 接口：在 `api/app/<领域>/` 增加 Handler → 在 `api/routes/*.go` 注册 → 必要时加中间件。
- 新增数据表：在 `internel/database/model/` 增加模型并迁移（项目若使用自动迁移则在启动或独立命令中完成）。
- 本文若与代码不一致，以当前代码为准；`api/app` 下文件较多，未逐文件列名的 `.go` 均可按「单文件单 Handler」理解。
