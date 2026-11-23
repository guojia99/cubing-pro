# 配置参考

本页按 `src/configs/` 中的 YAML 标签及实际消费者整理（加载器和公共配置在 `configs.go`，全局、网关、机器人结构分别在 `global.go`、`gateway.go`、`robot.go`）。公开的 [`etc/server.yaml`](../etc/server.yaml) 是本地开发示例；字段缺省值与示例值不同的地方在下文注明。它不提供完整离线环境，也不保证未配置依赖的功能可用。生产上线前的配置核对、材料清单和验收步骤见[后端上线须知与材料清单](production-go-live.md)。

## 加载与路径

```sh
make init-config
make config-test
# 确认依赖和后台行为后再按需启动：
make api
make gateway
# 覆盖 Makefile 的配置选择：
make api CONFIG=/absolute/path/config.yaml
```

Makefile 的 `CONFIG ?= ./local/server_local_dev.yaml` 保留旧默认值，应用 CLI 自身的 `--config` / `-c` 默认却是 `./etc/server.yaml`。`make init-config` 仅在目标不存在时复制示例，不覆盖或合并已有文件；修改示例不会更新私人副本或生产配置。`CONFIG` 只用于 Makefile，不是应用的环境变量配置接口。

加载器读取文件后用 `yaml.v3` 解码，没有环境变量替换、自动读取 `.env` 或按环境覆盖字段的逻辑。`"${PASSWORD}"` 会作为字面字符串读取，`""` 只是空值。没有严格拒绝未知字段，所以拼错键可能被静默忽略。除下述消费者默认值外，缺失的字符串、数字、布尔、列表分别为 `""`、`0`、`false`、空列表。

所有相对文件路径以**进程当前工作目录（cwd）**为基准，不以配置文件所在目录为基准；把 YAML 移到 `local/` 不会改变 `./logs` 的位置。`make` 在仓库根目录运行时，本地路径指向仓库；现有 systemd 文件的 cwd 为 `/usr/local/bin`。部署应使用绝对文件路径。URL 路径（如动态页 `placeholder`）和明确相对于资源根目录的子路径不属于 cwd 文件路径。

## 日志 `log`

| 字段 | 行为 / 默认 |
| --- | --- |
| `path` | CLI 初始化服务后传给滚动日志器的目录；固定文件名 `cubing-pro.log`。示例 `./logs`，加载器不补默认目录。生产环境应使用绝对路径，例如 `/var/log/cubing-pro`。 |
| `maxSize` | 滚动日志大小，传入单位 MB；示例 `100`。加载器不补默认大小。 |

须保证目录可写。相对路径按进程 cwd 解析，不会以 YAML 所在目录为基准。配置只影响该滚动日志器，不能控制所有标准输出或 QQ webhook 的 `./logs` 日志。

## 全局 `global`

| 字段 | 行为 / 默认 |
| --- | --- |
| `baseHost` | 对外 URL 根地址，邮件链接和 WCA OAuth 回调地址的后备值，不决定监听地址。示例 `http://127.0.0.1:8080`，生产需自行设置。 |
| `debug` | 启用 MySQL SQL 调试日志；也会跳过账号密码登录中的验证码检查。示例 `true`，仅用于受控本地开发；缺省 `false`。不是统一的功能或网络开关。 |
| `db.driver` | 主业务库唯一支持值为 `mysql`，其他值或空值会报错；不提供 SQLite 回退。 |
| `db.dsn` | MySQL DSN，形如 `USER:PASSWORD@tcp(HOST:3306)/cubing_pro?charset=utf8mb4&parseTime=True&loc=Local`。须提供可用值，服务初始化会自动迁移业务表。不是 WCA 数据库配置。 |
| `algTrainersPath` | 非空时启动期间调用公式库初始化；空值跳过，公式功能缺少数据。路径及约束见下文。 |

### 打乱 `global.scramble`

| 字段 | 行为 / 默认 |
| --- | --- |
| `type` | 消费者接受 `tnoodle` / `rust_twisty`，空值默认 `tnoodle`。历史注释中的 `lang` 不是实现支持的类型，未知类型不会自动回退。 |
| `endpoint` | tNoodle 兼容 HTTP 服务根地址，访问 `/frontend/data/events`、颜色和打乱接口。示例 `http://127.0.0.1:20009`，不会自动启动服务，也没有缺省 endpoint。 |
| `scrambleDrawType` | 空值默认 `2mf8`，目前绘图分支只实现 `2mf8`，会访问代码写定的外部绘图服务。 |
| `scramble` | 可解码并保存到打乱对象，但当前生成及绘图代码没有读取它，不会覆盖绘图服务 URL。 |

默认 CLI 会创建打乱组件，没有可用的 `--scramble` 开关（定义已注释）。`rust_twisty` 在非 macOS 构建需要相应 C/Rust 库；macOS 实现返回占位内容，不能作为可用打乱后端。选择 tNoodle 也不能绕过非 macOS 构建中的链接依赖。`TNoodleScrambles` 的整场生成入口当前仍是未实现的 panic，不能据示例宣称比赛打乱功能完整。

### 邮件 `global.emailConfig`

| 字段 | 行为 / 默认 |
| --- | --- |
| `smtpHost` | SMTP 服务器地址。 |
| `smtpPort` | SMTP 端口，直接传给邮件客户端。 |
| `from` | 发件地址，也是 SMTP 认证用户名。 |
| `fromName` | 发件显示名，示例 `cubingPro`。 |
| `password` | SMTP 密码 / 授权码，只放私人副本。 |

没有邮件 `enable` 开关，空值不会使注册、找回密码、主办邀请等流程自动跳过发信；调用发信时会失败，需要单独配置 SMTP。

### WCA 数据 `global.wcaDB`

键名区分写法，必须是 `wcaDB`，不是旧示例的 `wca_db`。

| 字段 | 行为 / 默认 |
| --- | --- |
| `mysqlUrl` | 独立 WCA MySQL DSN 前缀，形如 `USER:PASSWORD@tcp(HOST:3306)/`，由代码拼接数据库名及查询参数。不要填写实际凭据到公开文件；不能用主业务库配置替代独立 WCA 库配置。 |
| `dbPath` | WCA 解压数据及 `wca.txt` 数据库名指针所在目录；不是 SQLite 文件。空值时指针读取位置是 cwd 的 `wca.txt`。 |
| `syncPath` | WCA 下载 ZIP 的暂存目录。启用同步前需配置为可写目录。 |

三项空值表示示例未准备 WCA 数据，不是全局禁用开关。构造 WCA 客户端时仍尝试读取本地 `wca.txt`；不带 `-w` 时每小时重读指针，并在有指针时尝试连接 MySQL。缺少数据会记录失败日志，相关查询不可用或可能出错。

`--sync-wca` / `-w` 默认 `false`。开启后会立即进入同步循环并每小时执行一次，其中下载 / 导入只在上海时间 `[01:00, 09:00)` 进行，之后刷新数据库连接。代码使用固定的 WCA SQL 导出地址，需要网络、MySQL、`mysql` 命令及相关导入工具；会创建、删除 `wca_YYYYMMDD` 数据库并清理旧数据，须隔离账号权限及存储，且只由一个实例承担。`job.unRunJobs` 不过滤此循环。不要对空配置执行 `make api_w`。

### WCA 登录 `global.wcaAuth2`

| 字段 | 行为 / 默认 |
| --- | --- |
| `appId` | WCA OAuth Client ID；与 QQ 的 `app_id` 不是一个字段。 |
| `appSecret` | WCA OAuth Client Secret；任一凭据为空时登录处理器返回未配置错误。 |
| `redirectBase` | 回调 URL 根地址；为空回退到 `global.baseHost`，追加 `/v3/cube-api/auth/wca/callback`。需与 WCA 应用登记地址一致。 |
| `frontendBase` | 允许跳转来源之一，正常跳转分支的后备根地址优先于 `redirectBase`。实际目标来自请求的 `redirect`；未传或为 `debug` 时回调返回 JSON，不自动跳转前端。 |
| `redirectURLs` | 非空列表优先作为允许跳转来源；否则从 `frontendBase`、`redirectBase` 收集。不会替代 OAuth 回调根地址。 |
| `auths` | OAuth scopes，空列表默认 `public dob email openid profile`。 |
| `authsPath` | 仅声明，当前 OAuth 消费者未读取。 |

需前端登录流程、可用业务数据库及 WCA 网络访问；本地示例的 URL 需要按实际前端入口修改，空凭据不影响 API 路由注册但不能登录。

### 公式目录

例如设置 `global.algTrainersPath: "/srv/cubing-pro/formulas/Alg-Trainers"` 时，必须同时准备 `/srv/cubing-pro/formulas/Algs/catalog.json` 及 catalog 引用的文件。代码从 `filepath.Dir(algTrainersPath)/Algs` 先加载公式库，catalog 不存在、为空、条目非法或数据不完整会使服务初始化失败；之后才从 `Alg-Trainers` 读取补充集合，个别补充集合读取失败会跳过。不能只复制 `Alg-Trainers` 或把 `Algs` 放到它内部。

## API `api`

| 字段 | 行为 / 默认 |
| --- | --- |
| `host` | API 监听地址，同时是网关 API 反代的目标 host；示例保留 `localhost`。生产若只经网关访问，建议使用回环地址；若设为通配地址，需确认反代目标在部署环境可达，并限制直接访问。 |
| `port` | API 监听端口及网关目标端口；示例保留 `20000`。没有加载器默认端口，`0` 会由监听器分配临时端口，不能作为稳定代理目标。 |
| `staticPath` | 上传资源根目录，头像放在 `avatar/`，广告图放在 `advertisements/`；非空时 API 尝试创建根目录。空值使用 cwd 下的子目录。示例 `./local/static`，不是前端产物目录，也不会把整个目录自动挂载成网关静态站点。 |

## 网关 `gateway`

### 监听与代理

| 字段 | 行为 / 默认 |
| --- | --- |
| `httpPort` | HTTP 监听端口，优先于旧 `port`；两者最终均为零时使用 `80`。示例明确 `8080`。 |
| `port` | 兼容旧字段，仅在 `httpPort == 0` 时生效。 |
| `httpsPort` | 只有大于零且 `pem`、`privateKey` 均非空时才开启 TLS。示例 `0`。 |
| `pem` / `privateKey` | TLS 证书与私钥路径；示例均空。TLS 启用后 HTTP 监听默认 301 重定向到 HTTPS。证书无效会导致 TLS 启动失败。 |
| `httpsHost` | HTTPS 重定向使用的主机名；空时从请求 Host 去除端口推导，非 443 HTTPS 端口自动附加。不是监听 host。 |
| `xFile` | `/v3/x-file` 文件根目录，示例 `./local/x-file`。不同于已失效的 `global.xFilePath`。 |
| `tNoodlePort` | 额外 tNoodle 代理的 localhost 上游端口；不影响 `global.scramble.endpoint`。 |
| `outsizeTNoodlePort` | 拼写按标签保留；为 `0` 时不启动额外代理监听。非零时另开 HTTP 端口；JS 内容替换仍硬编码旧域名与端口，部署前检查 `src/gateway/tnoodle.go`。 |
| `blddbPort` | 固定 Host `blddb.cubing.pro` 的 localhost 代理上游端口。缺省 `0` 不会移除该 Host 分支，也不意味着服务可用。 |

网关监听 `:<port>`，即所有网卡，没有可配置的监听 host。本地 HTTP 示例仍需控制外部访问。TLS 模式中 ACME challenge 路径及带 `X-Forwarded-Proto: https` 的请求可绕过 HTTP 重定向，部署前需评估前置代理与可信请求头。

路由优先级为固定 BLDDB Host、`localProxies`、`staticSites`、包含 `/v3/cube-api` 的 API 代理、包含 `/v3/x-file` 的文件路由、默认前端静态站。Host 站点或代理会接管整个 Host，包括其 API 路径。API 代理目标来自 `api.host` / `api.port`。

### 默认前端静态资源

| 字段 | 行为 / 默认 |
| --- | --- |
| `staticRoot` | 非空时最高优先，整站目录模式覆盖 `indexPath` / `staticPath`，入口 `index.html`。示例 `./local/frontend/out`，须自行准备产物。 |
| `indexPath` | 无 `staticRoot` 时的入口文件。`staticPath` 空且 `spa: false` 时，以入口所在目录按多页站点提供文件。 |
| `staticPath` | 遗留 Umi 分支的资源目录；只在未使用 `staticRoot`、且不走上述多页模式时参与扩展名匹配。不是 `api.staticPath`。 |
| `spa` | 整站目录模式中，物理文件未命中且 URL 无扩展名时，`true` 回退入口，`false` 尝试动态占位页再 404。缺省 `false`，Next 多页导出应保持 `false`。遗留 `indexPath` + `staticPath` 分支的其余请求仍回退入口，不由此字段关闭。 |
| `dynamicRouteFallbacks[].match` | Go 正则表达式，按列表顺序匹配；空值或非法正则跳过。仅整站 / 多页模式适用。 |
| `dynamicRouteFallbacks[].placeholder` | 相对于静态根目录的预构建 HTML 路径，可带起始 `/`；文件必须存在。`spa: true` 的入口回退优先于动态占位页，URL 有扩展名且物理文件未命中时直接 404。 |
| `staticFileExts` | 仅遗留 `staticPath` 分支使用；非空列表替换默认列表，匹配带点扩展名且区分大小写。默认 `.css .js .svg .webp .woff .png .jpeg .jpg .ico .json .md`。 |

示例保留 competition、player、wca/player 和 admin organizers result 四组动态规则。后端不会生成前端页面，构建产物中必须有对应 `__dynamic__` 文件；没有前端产物时静态请求可能为 404，不代表 API 已启动失败。整站 / 多页与遗留分支的默认缓存头均为 `public, max-age=60`。

### 按 Host 的站点 `gateway.staticSites[]`

| 字段 | 行为 / 默认 |
| --- | --- |
| `host` / `hosts` | 非空 `hosts` 列表优先于 `host`，不会合并（列表中全是空字符串也不会回退到 `host`）。host 键去首尾空白并转小写，请求 Host 去端口；配置应写无端口域名。相同 host 后配置覆盖前配置。 |
| `root` | 静态根目录，空值跳过站点；相对路径按 cwd。 |
| `index` | 相对于 root 的入口名，空值默认 `index.html`；子目录查找使用入口 basename。 |
| `spa` | 未命中物理文件且 URL 无扩展名时是否回退入口，缺省 `false`。 |
| `dynamicRouteFallbacks` | 该站点自己的 `match` / `placeholder` 列表，不自动继承默认站点的规则。 |
| `cacheControl` | 响应缓存头，空值默认 `public, max-age=60`。 |
| `autoUpdate` | 缺省 `false`。配合 CLI `-j` 注册 Git 自动更新任务；实际注册在共享服务初始化中，不仅限于 api 子命令。 |
| `autoUpdateInterval` | Go duration 如 `5m`、`1h`，空、非法或非正数使用 `5m`。 |
| `repoDir` | Git 和构建命令的工作目录，空值用 `root`；相对路径仍按进程 cwd。产物目录不是仓库根时必须填写。 |
| `buildCmd` | HEAD 因 `git pull` 变化后用 `/bin/sh -c` 执行，空值默认 `npm run build`；继承进程环境。不是 YAML 环境变量插值。需自行安装 Git、Node 和前端依赖。 |

任务启动时立即执行一次，之后按间隔执行；部署时由一个实例更新同一仓库。`root` 非空才注册自动更新任务，`repoDir` 单独设置不能替代 `root`。

### 按 Host 的代理 `gateway.localProxies[]`

| 字段 | 行为 / 默认 |
| --- | --- |
| `host` / `hosts` | 列表优先、规范化及重复覆盖规则与 staticSites 相同。 |
| `port` | 上游 HTTP 端口，必须大于零，否则跳过条目。 |
| `targetHost` | 上游地址，空或全空白默认 `127.0.0.1`。只能配置该 HTTP 代理目标，不是含 scheme 的完整 URL。 |

## 机器人 `robot`

所有机器人列表默认空；公开示例逐类提供禁用条目。只有 robot 子命令构造实例，API 或 gateway 不会顺带启动机器人。`robot.personValPath` 是人物素材插件的根目录，但当前 `NewPlugins` 未注册该插件，单独配置路径不会启用它。[机器人说明](../src/robot/README.md) 介绍运行依赖。

### QQ `robot.QQBot[]`

| 字段 | 实际行为 |
| --- | --- |
| `enable` | 唯一的实例创建开关，默认 `false`。 |
| `qq` | uint64 QQ 号码，传入内置 webhook 应用设置，示例 `0`。 |
| `app_id` | uint64 应用 ID，用于 token、OpenAPI 和 WebSocket 连接；不是 `appID` / `appId`。 |
| `token` | 传入内置 webhook 应用设置；当前 access token 使用 `app_id` + `app_secret` 获取，不直接使用该字段。 |
| `app_secret` | 用于 access token 获取、鉴权连接及 webhook 处理。 |
| `is_sandbox` | 决定客户端 OpenAPI 沙箱及认证记录；但 webhook 应用设置中的沙箱值当前硬编码为 `false`，不是全链路沙箱保证。 |
| `wss_addr` | 显式 WebSocket 地址，无消费者默认值。需与 webhook 的端口、TLS 和模式匹配。 |
| `is_open` | `true` 使用 `ConnectUniversalWithSecret`，`false` 使用 `ConnectUniversal`，都尝试连接。不是启用开关。 |
| `server.port` | 内置 webhook HTTP(S) 服务端口。`0` 是临时端口，不能直接与固定 WebSocket 地址配套；失败时实现还可能尝试 80/8080/443/8443。 |
| `server.certFile` / `server.certKey` | 两项均非空才运行 webhook TLS，否则 HTTP；与 gateway 的 TLS 配置独立。 |
| `server.isOpen` | 选择 webhook 路由模式：`false` 提供 `/websocket` 及 `/qqbot/:appid`；`true` 提供 `/wss/qqbot` 及带 app secret 的回调路由。两个值都会启动监听。 |

QQ 启用后同时启动内置 webhook、向 QQ 获取 token，并连接配置的 WebSocket。需要有效平台应用及权限、可达回调入口和正确的模式配置；零 / 空占位不能直接启用。内置 webhook 使用进程级全局设置，多 QQ 实例及端口共存需额外验证，不承诺配置多个条目即可独立运行。

### CQHTTP `robot.CQHttpBot[]`

| 字段 | 实际行为 |
| --- | --- |
| `enable` | 是否创建 CQHTTP 实例，缺省 `false`。 |
| `prefix` | 声明为命令前缀，适配器的 `Prefix()` 会返回它，但当前消息分发没有调用此方法，不会据此自动过滤或去掉前缀。 |
| `address` | 外部 CQHTTP HTTP API 根地址，发消息会追加 `/send_group_msg`。 |
| `post` | 拼写就是 `post`，整数反向 HTTP 接收端口；监听 `127.0.0.1:<post>`，不是 URL 或 `port` 键。 |

需另行部署兼容 CQHTTP 服务并把群消息事件回送到接收端口，不会自动安装或登录 go-cqhttp。CQHTTP 图片发送使用 `file://` 路径，外部服务必须能读取相应文件。

### WeChat `robot.WeChatBot[]`

只有 `enable` 字段被声明，当前机器人创建流程未读取该列表，WeChat 未接通；设置为 `true` 不会启动微信机器人。

## 定时任务 `job` 与初始化边界

`job.unRunJobs` 按精确任务名排除注册项。基础任务为 `RecordUpdateJob`（30 分钟）及 `UpdateCubingChinaComps`（24 小时，涉及外部数据抓取）。`UpdateDiyRankings` 当前未注册。站点任务名为 `StaticSiteAutoUpdate:<StableID>`，ID 优先使用 `hosts` 中首个非空值（小写、去空白），再用 `host`，最后用 `root`。

- `--job` / `-j` 默认关闭，启用后启动已注册任务循环，任务立即执行一次再按间隔执行。`unRunJobs` 过滤在注册时发生，也影响按名称手动调用这些任务；不是整个程序的副作用开关。
- `--sync-wca` / `-w` 默认关闭，与 `-j` 相互独立。`make api_j` 仅加 `-j`，`make api_w` 仅加 `-w`，均不要对未准备的生产依赖盲目运行。
- API、gateway、robot 和 admin 等子命令共享服务初始化，都可能打开业务库并执行迁移、创建打乱组件、加载公式、创建 WCA 客户端；不是只初始化自身那一组配置。admin 会写入管理员记录，不能用作配置检查命令。
- Seniors 在 `src/internel/wca_api/wcaSeniors.go` 的包级 `init()` 中独立启动 goroutine：读 cwd 下 `wca_seniors_cache.json`，缓存不可用 / 过期时请求外部 Seniors 数据，并每 6 小时更新。此行为不受 `-j`、`-w`、`job.unRunJobs`、WCA 空配置或机器人开关控制，部分测试或 CLI 路径也可能因导入该包触发。
- `make config-test` 只执行 `go test ./src/configs`，当前覆盖公开 `etc/server.yaml` 的严格字段解码、网关兼容优先级及加载错误；不创建共享服务、不检查 MySQL / SMTP / tNoodle / 证书 / 前端产物，也不验证任意 `CONFIG` 文件。

## 兼容与迁移

| 旧配置 / 易混字段 | 当前规则 / 操作 |
| --- | --- |
| 顶层 `apiGateway` | 仅当顶层不存在 `gateway` 键时作为别名使用；不是字段级合并。即使 `gateway` 是空映射或 null，也不会从旧键补齐。新配置只写 `gateway`。 |
| `gateway.port` | 仅当 `httpPort` 为零时迁移为 HTTP 端口；非零 `httpPort` 优先。 |
| `global.xStaticPath` | 无对应有效字段，移除。API 上传资源配置到 `api.staticPath`，前端资源使用 gateway 的静态字段；两者不能机械等价替换。 |
| `global.xFilePath` | 不消费；如需 `/v3/x-file` 路由，迁移到 `gateway.xFile` 并核对路径与访问范围。 |
| `global.algPath` | 无对应有效字段，移除。公式库位置只由 `global.algTrainersPath` 控制，且要求同级存在 `Algs/catalog.json` 及引用文件。 |
| `global.wca_db.syncUrl` / `syncTime` | 不消费，不能简单改名延续旧行为。改用 `global.wcaDB` 的三项实际字段，下载来源 / 调度在实现中固定，开启同步仍需 `-w`。 |
| QQ `group` / `appID` / `groupList` | 没有这些 YAML 标签；移除，ID 改为 `app_id`，其余配置按实际 QQ 字段补齐。没有等价的群白名单配置。 |
| 生产 `/build`、`/https`、80/443 | 只属于部署环境约定。新版公开示例使用本地目录及 HTTP 8080，已有私有 / 生产副本需按实际环境保留或调整，不能直接整份替换。 |

其余仅声明或未接通项：`global.dev`、`global.imageTempPath`、`global.baseFontTTf` 当前没有业务消费者；`global.wcaAuth2.authsPath` 未读取；`global.scramble.scramble` 只保存未使用；`robot.CQHttpBot[].prefix` 未在消息分发中使用；`robot.personValPath` 的插件未注册；`robot.WeChatBot` 未接入创建流程。不要将这些字段写进配置后就认为相关功能已启用。

迁移前备份私人配置与数据库，逐项核对差异；新示例、`make init-config`、`make install` 都不会自动改写私人 YAML。配置、功能依赖或优先级有变化时，请同步贡献示例、本页、[etc 说明](../etc/README.md) 及相关消费者测试，避免提交真实凭据。
