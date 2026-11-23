# 后端上线须知与材料清单

本文面向使用 `/etc/cubing_pro/config.yaml` 和仓库 `etc/*.service` 的 Linux/systemd 部署，说明当前后端启动链、上线前必须核对的配置与材料，以及启动后的验收和回滚边界。

本文只引用配置键名，不记录数据库、OAuth、QQ、SMTP 等真实凭据。实际路径、域名、端口和资源以目标服务器为准；`local/prod.yaml` 是脱敏后的线上配置，修改前应先备份。

通用字段说明见[配置参考](configuration.md)，systemd 安装和更新见[部署说明](deployment.md)。

## 1. 启动链

API、Gateway、Robot 都通过 `cmd/root.go` 的共享初始化：

1. 读取 `--config` 指定的 YAML。
2. 连接业务数据库。
3. 构造打乱组件、WCA 客户端和业务服务。
4. 对业务表执行 AutoMigrate。
5. 设置滚动日志。
6. 进入具体子命令。

因此，Gateway 和 Robot 不是轻量静态服务：业务 MySQL 不可用时，它们也会启动失败。迁移错误在当前实现中被忽略，不能把进程启动成功当作数据库结构完整。

`api` 子命令还会在监听 HTTP 之前执行公式库初始化。数据库尚未完成首次导入时，缺少有效的 `Algs/catalog.json` 或其引用文件会使 API 启动失败；已标记导入完成时，本次启动会跳过公式源文件。

systemd 文件的固定约定：

| 项目 | 当前值 |
| --- | --- |
| 二进制 | `/usr/local/bin/cubing-pro` |
| 配置 | `/etc/cubing_pro/config.yaml` |
| 工作目录 | `/usr/local/bin` |
| 子命令 | `api`、`gateway`、`robot` |
| 运行用户 | 未设置 `User` / `Group`，系统服务默认 root |
| 任务参数 | 三个 unit 均未带 `-j` / `-w` |

## 2. 上线前须知

### 2.1 配置加载

- 加载器只读取一个 YAML 文件，不做环境变量插值、`.env` 读取或多文件合并。
- 未知键可能被静默忽略；字段拼错不会阻止启动，但对应功能不会生效。
- 所有相对文件路径按进程 cwd 解析，而不是按 YAML 所在目录解析。现有 unit 的 cwd 是 `/usr/local/bin`，生产配置应使用绝对路径。
- `make install` 不会安装或合并配置，也不会准备外部资源。

### 2.2 监听地址与暴露面

| 监听项 | 行为 |
| --- | --- |
| Gateway HTTP/HTTPS | 监听 `:<httpPort>` / `:<httpsPort>`，即所有网卡，没有 bind host 配置。启用 HTTPS 需同时满足端口为正且证书、私钥均非空。 |
| API | 监听 `api.host:api.port`。`api.host` 同时被 Gateway 当作 API 反代目标，一个字段承担监听和上游两种语义。 |
| QQ webhook | 监听 `server.port`，即所有网卡；`server.isOpen` 只选择路由模式，不是关闭开关；`server.port: 0` 会申请临时端口，也不是关闭。监听失败时实现还可能尝试其他内置端口。 |
| 额外 tNoodle 代理 | `gateway.outsizeTNoodlePort` 非零时另开一个全网卡明文 HTTP 监听，上游是 `localhost:tNoodlePort`。 |
| BLDDB / localProxies | 按 Host 反代到 `localhost:blddbPort` 或 `targetHost:port`，不是独立监听端口。 |

建议：

- 如果 API 只允许通过 Gateway 访问，将 `api.host` 改为回环地址，避免 20000 端口直接暴露；同时确认防火墙没有放行该端口。
- 只对外开放 80/443，以及机器人回调确实需要的端口。`outsizeTNoodlePort`、API 端口和内部代理端口应按实际需要限制来源。
- Gateway 的 HTTPS 重定向会信任带 `X-Forwarded-Proto: https` 的请求；若前面有反向代理或 CDN，需确认可信转发头范围。

### 2.3 首次启动副作用

- `api` 可能执行公式库首次导入，耗时和失败都发生在开始监听前。
- 所有子命令都会尝试 AutoMigrate 业务表，部分错误不会阻止启动；首次上线或结构变更后应检查日志和表结构。
- WCA 客户端启动后会尝试读取 `wcaDB.dbPath` 下的 `wca.txt`，并每小时刷新连接。指针缺失或指向的 MySQL 不可用不会阻止进程启动，但 WCA 查询不可用。
- Seniors 相关包的 `init()` 会读取 cwd 下的 `wca_seniors_cache.json`，缓存缺失或过期时访问外部 Seniors 服务，并每 6 小时更新。该行为不受 `-j`、`-w` 或机器人开关控制。
- `-j` 会启动已注册任务循环，任务在启动时立即执行一次。静态站点 `autoUpdate` 也要求 `-j`，并会执行 `git pull` 和构建命令。
- `-w` 会启动 WCA 同步循环，可能下载、解压、创建或删除 `wca_YYYYMMDD` 数据库，并调用 `mysql`、`sed`。同一 WCA 数据目录和 MySQL 实例只应有一个同步实例。

### 2.4 已知代码级限制

以下问题本次不通过配置修复，只在上线时记录和规避：

- JWT 签名密钥当前硬编码在代码中，没有 YAML 配置入口。轮换密钥需要代码改动，并会使旧 token 失效。
- Gateway、extra tNoodle 代理和 QQ webhook 都没有 bind host 配置，默认监听所有网卡。
- `wca_seniors_cache.json` 固定使用进程 cwd，不随 `log.path` 或配置目录变化。
- `global.imageTempPath`、`global.baseFontTTf` 当前没有业务消费者；字体通过 Go embed 编入二进制。
- `global.scramble.scramble` 只保存到打乱对象，当前生成和绘图代码没有读取它；绘图服务地址仍是代码常量。
- `robot.personValPath` 对应的人物素材插件当前未注册；`robot.WeChatBot` 未接入启动流程；CQHTTP 的 `prefix` 未在消息分发中使用。
- `algdb/` 目录不是当前公式库路径；`HuaWenHeiTi.ttf` 放在 `/etc/cubing_pro` 也不是当前运行时的必需文件。

## 3. 线上配置核对

以下按 `local/prod.yaml` 的实际键整理。不要直接把本文当作配置模板覆盖线上文件；逐项确认后再修改。

| 配置键 | 上线建议 |
| --- | --- |
| `log.path` | 改为绝对路径，建议 `/var/log/cubing-pro`。创建目录并授予实际运行账号写权限；文件名固定为 `cubing-pro.log`。 |
| `log.maxSize` | 按磁盘容量设置滚动大小；日志目录不可写时不能依赖该配置。 |
| `global.algPath` | 当前无对应结构体字段和消费者。删除或注释；公式库位置由 `global.algTrainersPath` 控制。 |
| `global.algTrainersPath` | 保留实际部署目录，并确保同级存在 `Algs/catalog.json`。只部署 `Alg-Trainers` 不够。 |
| `global.db` | 确认 MySQL 地址、账号权限和数据库备份。生产账号使用最小权限，不使用开发或共享高权限账号。 |
| `global.wcaDB.mysqlUrl` | WCA 独立 MySQL DSN 前缀；确认指向 WCA 数据实例，不要与业务库混用。 |
| `global.wcaDB.dbPath` | `wca.txt` 指针所在目录，例如 `/etc/cubing_pro/wca/sync_path`；确认指针文件和目标数据库存在。 |
| `global.wcaDB.syncPath` | 仅在启用 `-w` 时需要可写；准备下载、解压和清理所需磁盘空间。 |
| `global.scramble.type` | 当前线上使用 `rust_twisty`；非 macOS 构建必须链接匹配的 `librust_scramble`。若改用 `tnoodle`，需另备外部服务。 |
| `global.scramble.endpoint` | 仅 `tnoodle` 类型直接使用；`rust_twisty` 不依赖该 endpoint，但 Gateway 的 tNoodle 代理仍可能使用 `tNoodlePort`。 |
| `global.scramble.scrambleDrawType` | 当前绘图分支只实现 `2mf8`，访问代码内固定的外部绘图服务；确认出网和可用性。 |
| `global.emailConfig` | 没有 enable 开关；SMTP 为空时发信流程会在调用时失败。确认主机、端口、账号授权码和发件地址。 |
| `global.wcaAuth2` | 确认 WCA 应用凭据、回调根地址、前端允许来源和 scopes；回调路径为 `/v3/cube-api/auth/wca/callback`。 |
| `gateway.httpPort` | 当前使用 80；若前置代理终止 HTTP，可改为非特权端口并调整入口。`httpPort: 0` 最终会回落到 80，不是关闭。 |
| `gateway.httpsPort` | 大于 0 且 `pem`、`privateKey` 均非空才启用内置 TLS；不需要内置 TLS 时显式设为 0。 |
| `gateway.pem` / `gateway.privateKey` | 确认文件存在、证书链完整、域名匹配、未过期、私钥与证书配对，且运行账号可读。 |
| `gateway.httpsHost` | 只用于 HTTPS 重定向主机名，不是监听地址；多域名或前置代理场景需核对。 |
| `gateway.staticRoot` | 主站静态产物根目录；必须有 `index.html` 和四个动态占位页。后端不会生成或复制前端文件。 |
| `gateway.dynamicRouteFallbacks` | 逐条确认 `placeholder` 文件存在；缺少时对应动态 URL 会 404。 |
| `gateway.staticSites` | 逐站点确认 `root`、`index`、`repoDir` 和构建产物。`autoUpdate: true` 只有在带 `-j` 的实例上才生效。 |
| `gateway.xFile` | `/v3/x-file` 的实际文件根目录；确认目录存在、读取权限和访问范围。 |
| `gateway.tNoodlePort` | extra tNoodle 代理的 localhost 上游端口；对应服务不存在时相关代理请求失败。 |
| `gateway.outsizeTNoodlePort` | 不需要额外 tNoodle 代理时设为 0；需要时确认明文监听、防火墙和 JS 替换中的硬编码域名/端口。 |
| `gateway.blddbPort` | 固定 Host `blddb.cubing.pro` 的上游端口；缺省 0 不会移除该 Host 分支，只会让请求失败。 |
| `gateway.localProxies` | 确认每个 `host`、`targetHost`、`port` 都可达；这些是内网代理入口，需审查访问范围。 |
| `api.host` | 同时是监听地址和 Gateway 上游。若 API 只经 Gateway 访问，建议改为 `127.0.0.1`；若设为 `0.0.0.0`，需确认 20000 端口没有直接暴露。 |
| `api.port` | API 监听端口和 Gateway 上游端口；`0` 会分配临时端口，不能作为稳定代理目标。 |
| `api.staticPath` | 建议保留 `/data/static`。API 会在非空时尝试创建目录，但创建错误被忽略；需提前确认目录存在、可写，并备份 `avatar/`、`advertisements/`。该目录不是 Gateway 静态站点。 |
| `robot.personValPath` | 当前插件未注册，配置不会启用人物素材功能。建议删除或改为真实目录，并在功能接入后再启用。 |
| `robot.QQBot[].enable` | 唯一的实例创建开关。启用前确认平台应用、`app_id`、`app_secret`、`wss_addr`、回调入口和证书。 |
| `robot.QQBot[].is_open` | 选择 WebSocket 连接方式，不是启用开关；两种值都会尝试连接。 |
| `robot.QQBot[].server.port` | webhook 监听端口；`0` 是临时端口，不能与固定 `wss_addr` 稳定配套。 |
| `robot.QQBot[].server.isOpen` | 选择 webhook 路由模式，不是关闭开关；两种值都会启动监听。 |
| `robot.QQBot[].server.certFile` / `certKey` | 两项均非空才运行 webhook TLS；确认与 Gateway 证书独立配置且文件可读。 |
| `job.unRunJobs` | 只在 `-j` 生效时过滤已注册任务；不能关闭 `-w` 同步，也不能阻止 Seniors 初始化。 |

凭据类字段（数据库、WCA OAuth、QQ、SMTP）不在本文展开。上线前应确认最小权限、轮换计划和备份权限；`config.yaml` 建议权限 `0600`，`config.yaml.bck` 中的旧配置也应按含凭据文件处理。

## 4. 材料清单

### 4.1 程序与 systemd

- [ ] 已构建目标平台二进制，确认 Go 版本、CGO、C 编译器和 `librust_scramble` 链接条件满足。
- [ ] 已在目标机执行 `ldd /usr/local/bin/cubing-pro` 或等效检查，确认动态库可解析。
- [ ] 已备份旧二进制，并安装 `/usr/local/bin/cubing-pro`。
- [ ] 已安装 `cubing_pro_api.service`、`cubing_pro_gw.service`、`cubing_pro_robot.service`，并执行 `systemctl daemon-reload`。
- [ ] 已确认每个 unit 的 `ExecStart`、`WorkingDirectory`、运行账号和端口权限。
- [ ] 若改用专用账号，已通过 override 设置 `User` / `Group`，并授予配置读取、日志/数据目录写入和证书读取权限。

### 4.2 配置与证书

- [ ] `/etc/cubing_pro/config.yaml` 存在、YAML 有效、权限建议 `0600`。
- [ ] 已备份当前配置；`config.yaml.bck` 不作为运行配置，按旧凭据文件管理。
- [ ] 所有生产路径使用绝对路径，特别是日志、数据库、上传目录、公式库、WCA、证书和前端产物。
- [ ] 已确认 `api.host`、`api.port` 与 Gateway 上游一致，且不需要直接暴露 API。
- [ ] Gateway 证书和私钥文件存在、配对、未过期、域名匹配、证书链完整。
- [ ] QQ webhook 证书和私钥（如启用）存在、配对、可读，并与 `wss_addr` 的协议和端口匹配。
- [ ] 已确认 `gateway.outsizeTNoodlePort`、`blddbPort`、`tNoodlePort` 和 `localProxies` 的上游服务与暴露范围。

### 4.3 数据库

- [ ] 业务 MySQL 可连接，数据库存在，账号权限符合最小权限要求。
- [ ] 已完成业务库备份，并验证恢复流程。
- [ ] 首次启动后检查 AutoMigrate 日志；需要时按顺序评估并执行 `scripts/fix_automigrate_errors.sql`、`scripts/add_oauth_states_table.sql`、`scripts/add_wca_auth_columns.sql`。
- [ ] 手工 SQL 执行前已再次备份；不要把脚本当作无条件幂等迁移。
- [ ] WCA MySQL 可连接，`wca.txt` 指向的数据库存在且表完整。
- [ ] 如启用 `-w`，已准备可写 `global.wcaDB.syncPath`、足够磁盘空间、`mysql`/`sed` 工具和单独的同步实例。

### 4.4 公式库

- [ ] `global.algTrainersPath` 指向实际部署的 `Alg-Trainers` 目录。
- [ ] 同级存在 `Algs/catalog.json`，catalog 非空且条目合法。
- [ ] catalog 引用的每个路径都存在，且包含 `algs_info.json`、`algsets_info.json`、`groups_info.json`、`scrambles.json`。
- [ ] 已确认 `formula_catalog_imports.complete` 的当前状态；首次导入失败会阻止 API 启动。
- [ ] `algdb/` 仅作为历史资源核对，不作为当前公式目录。

### 4.5 WCA 数据

- [ ] `/etc/cubing_pro/wca/sync_path/wca.txt` 存在且内容为有效数据库名。
- [ ] `global.wcaDB.mysqlUrl` 指向的 WCA MySQL 可连接，账号有权读取目标库。
- [ ] 不需要同步时，三个 unit 均未带 `-w`；需要同步时只指定一个实例。
- [ ] 同步实例具备建库、删库、导入、清理权限，并已确认上海时间 `[01:00, 09:00)` 的下载导入窗口。
- [ ] 已确认 `wca/` 和暂存目录的写权限、磁盘空间和旧数据清理策略。

### 4.6 前端与静态子站

- [ ] 主站 `gateway.staticRoot` 下有 `index.html`。
- [ ] 四个 `dynamicRouteFallbacks` 占位页均存在：competition、player、wca/player、admin organizers result。
- [ ] `staticSites` 中每个站点的 `root`、`index` 和入口文件存在。
- [ ] `miniguildfold.cubing.pro` 对应的仓库、`dist` 和构建工具已准备；需要自动更新时，只在带 `-j` 的实例启用。
- [ ] `fmc-paperless`、`drm_doc` 等子站的构建产物已确认是否需要在本机提供。
- [ ] 已确认静态文件缓存头和动态路由回退行为符合预期。

### 4.7 上传、日志、缓存与临时目录

- [ ] `/data/static` 存在且运行账号可写；`avatar/`、`advertisements/` 可创建或已存在。
- [ ] 已核对数据库中旧图片的本地路径；如仍指向 `/etc/cubing_pro/static`，确认旧文件可读或制定迁移方案。
- [ ] `/var/log/cubing-pro` 存在且可写；`cubing-pro.log` 可正常滚动。
- [ ] 已确认 QQ webhook 在 cwd 下使用的 `logs/` 可写。
- [ ] 已确认 cwd 下 `wca_seniors_cache.json` 可写，或明确接受缓存写入失败。
- [ ] `/tmp` 有足够空间；WCA 同步和打乱图片等功能会使用临时文件。
- [ ] `/etc/cubing_pro/HuaWenHeiTi.ttf` 仅作为历史资源核对，当前二进制不依赖该路径加载字体。

### 4.8 外部服务与网络

- [ ] tNoodle 兼容服务（如使用 `tnoodle`）或本地 Rust 打乱库（如使用 `rust_twisty`）已准备。
- [ ] Gateway 的 extra tNoodle 代理所需上游（`tNoodlePort`）已启动，且 JS 替换域名/端口与实际一致。
- [ ] 打乱图片所需的 `2mf8` 外部绘图服务可达。
- [ ] SMTP 服务、发件账号和授权码可用。
- [ ] WCA OAuth 应用、回调地址和前端跳转允许列表与线上域名一致。
- [ ] QQ 开放平台应用、`app_id`、`app_secret`、回调入口和消息权限可用。
- [ ] WCA Seniors 数据源、WCA API 和赛事数据抓取所需网络可达。
- [ ] `mahjong.cubing.pro` 等 `localProxies` 目标、`blddb.cubing.pro` 上游和 DNS 解析可用。
- [ ] 防火墙只放行必要端口，并确认 20000、12014 等端口没有意外公网暴露。

### 4.9 备份与回滚

- [ ] 已备份业务数据库、配置、证书、上传文件、公式库和 WCA 指针。
- [ ] 已记录当前二进制版本和部署时间。
- [ ] 已准备回滚步骤：停止服务、恢复二进制/配置、按需恢复数据库和资源、启动并验收。
- [ ] 已知 AutoMigrate 可能修改表结构；回滚二进制前先评估数据库兼容性，不能承诺只换回二进制即可回滚。
- [ ] 已知 `make install` 会停止并重启三个服务，不会安装配置/unit，也不会选择性更新单个服务。

## 5. `/etc/cubing_pro` 目录对照

| 线上项 | 当前代码用途 | 上线动作 |
| --- | --- | --- |
| `config.yaml` | 三个 service 的主配置 | 按第 3 节核对；权限 `0600`。 |
| `config.yaml.bck` | 旧配置备份 | 不读取；按含凭据文件保护或清理。 |
| `Algs/` | 公式库 catalog 和数据 | 确认 `catalog.json` 及引用文件完整。 |
| `Alg-Trainers/` | 公式补充集合 | 与 `Algs/` 同级部署，确认被引用文件可读。 |
| `algdb/` | 历史公式目录 | 当前代码不读取；保留归档或确认无消费者后清理。 |
| `wca/` | WCA 指针/同步目录 | 确认 `sync_path/wca.txt`、WCA MySQL 和写权限。 |
| `ssl/` | Gateway/QQ 证书目录 | 确认 `cubing_pro_ca/cert.pem`、`key.pem` 实际存在且有效。 |
| `static/` | 当前配置未引用的历史目录 | 当前上传路径是 `/data/static`；核对旧图片引用和历史用途后决定迁移或保留。 |
| `HuaWenHeiTi.ttf` | 历史字体文件 | 当前字体已内嵌；不作为运行依赖。 |
| `log/` | 旧日志目录 | 若改用 `/var/log/cubing-pro`，确认旧日志归档后停用。 |
| `ohter/` | 子站资源 | 重点确认 `mini-guildfold-calculator` 的 `dist` 和仓库。 |
| `backup/`、`backup.sh` | 运维备份 | 当前启动不读取；核对是否覆盖数据库、上传、证书和配置。 |

## 6. 启动顺序与验收

### 6.1 只读预检

在目标机执行，按实际路径和账号调整：

```sh
sudo stat -c '%A %U:%G %n' \
  /etc/cubing_pro/config.yaml \
  /etc/cubing_pro/ssl/cubing_pro_ca/cert.pem \
  /etc/cubing_pro/ssl/cubing_pro_ca/key.pem \
  /etc/cubing_pro/Algs/catalog.json \
  /etc/cubing_pro/wca/sync_path/wca.txt \
  /data/static /var/log/cubing-pro

sudo namei -l /etc/cubing_pro/ssl/cubing_pro_ca/key.pem
sudo find /etc/cubing_pro/Algs /etc/cubing_pro/Alg-Trainers \
  -type f \( -name catalog.json -o -name algs_info.json \
  -o -name algsets_info.json -o -name groups_info.json \
  -o -name scrambles.json \) -print

sudo ss -lntp | grep -E ':(80|443|20000|8443|12014|2014|3000)\b'
sudo systemctl cat cubing_pro_api.service cubing_pro_gw.service cubing_pro_robot.service
```

数据库连通性建议使用只读账号或客户端配置文件，不要把密码直接写在命令历史里：

```sh
mysql --defaults-extra-file=/path/to/client.cnf -e 'SELECT 1'
```

### 6.2 启动顺序

1. 启动业务 MySQL 和 WCA MySQL（如使用本地 WCA 数据）。
2. 启动 tNoodle 或确认 Rust 打乱库可加载。
3. 启动 `cubing_pro_api.service`，确认公式初始化和监听成功。
4. 启动 `cubing_pro_gw.service`，确认 80/443 和 TLS 正常。
5. 按需启动 `cubing_pro_robot.service`，确认 webhook 和平台连接。

```sh
sudo systemctl start cubing_pro_api.service
sudo systemctl status cubing_pro_api.service
sudo journalctl -u cubing_pro_api.service -n 200 --no-pager

sudo systemctl start cubing_pro_gw.service
sudo systemctl status cubing_pro_gw.service
sudo journalctl -u cubing_pro_gw.service -n 200 --no-pager
```

### 6.3 验收项

- [ ] `journalctl` 中配置加载、数据库连接、日志初始化、公式导入和端口监听均成功。
- [ ] API 直连（若监听回环）和经 Gateway 的 `/v3/cube-api` 请求均返回预期结果。
- [ ] `https://cubing.pro/` 返回主站入口；证书链、域名和有效期正确。
- [ ] competition、player、wca/player、admin result 等动态 URL 命中占位页而不是 404。
- [ ] `miniguildfold.cubing.pro` 等子站入口正常；未启用 `-j` 时不期待自动更新。
- [ ] WCA 查询使用 `wca.txt` 指向的数据库；缺少数据时接口返回明确失败而不是被误判为服务正常。
- [ ] 上传功能只在授权测试中验证，并清理测试文件；不要在生产随意创建头像或广告图。
- [ ] QQ 机器人按实际平台流程发送测试消息；`wss_addr`、webhook 路由模式和证书一致。
- [ ] 业务数据库备份任务、日志滚动和磁盘空间告警已生效。

Gateway 返回 200 不代表 API、数据库、公式库、WCA 或机器人正常；应分别验证实际业务路径。

## 7. 回滚与记录

回滚前先停止相关服务，备份当前数据库和配置，再恢复上一版本二进制、配置和资源。由于启动时可能执行 AutoMigrate，回滚二进制前要评估表结构变化；必要时按备份恢复数据库。

建议记录：

| 日期 | 变更 | 执行人 | 验证结果 | 回滚点 |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |

`make install` 会构建二进制、停止三个服务、复制二进制并重启三个服务；它不会安装配置或 unit，也不会只更新单个服务。选择性更新应手动备份、停止选定服务、安装二进制并重启。

## 8. 相关文档

- [配置参考](configuration.md)：字段语义、默认值和迁移规则。
- [部署说明](deployment.md)：systemd 安装、权限、更新和回滚。
- [开发指南](development.md)：构建条件、启动链和测试边界。
- [脚本说明](../scripts/README.md)：公式数据导入和维护脚本。
