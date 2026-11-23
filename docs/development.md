# 本地开发与验证

本文说明开发入口及其副作用。配置字段见[配置说明](configuration.md)，上线操作见[部署指南](deployment.md)，模块位置见[源码导航](../src/SOURCE_STRUCTURE.md)。以下命令均从仓库根目录执行。

## 环境与首次配置

`go.mod` 要求 Go 1.25.0 起。准备 Make、启用 CGO 的 Go 环境和 C 编译器；SQLite 驱动依赖 `go-sqlite3`。非 macOS 的打乱包还会链接 `src/internel/scramble/librust_scramble.a`，库文件必须匹配目标系统与架构。

```bash
go mod download
make init-config
```

`init-config` 默认将 `etc/server.yaml` 复制到 `local/server_local_dev.yaml`，指定 `CONFIG` 时改用该目标路径，不会覆盖已有文件。公开示例使用本地 HTTP、SQLite，机器人默认禁用；检查路径、端口和依赖后再启动。不要直接使用包含线上地址或凭据的私人配置。相对文件路径按进程工作目录解析，不按 YAML 所在目录解析，因此本地命令应在仓库根目录运行。

Makefile 的 `CONFIG` 默认值为 `local/server_local_dev.yaml`；根 Cobra 命令自身的 `--config` 默认值则为 `./etc/server.yaml`。直接使用 Go 命令时应明确指定配置，避免混淆：

```bash
make api CONFIG=local/server_local_dev.yaml
# 等价的直接启动入口
go run main.go api --config ./local/server_local_dev.yaml
```

API、Gateway、Robot 是不同进程，按需在各自终端启动：

```bash
make gateway CONFIG=local/server_local_dev.yaml
make robot CONFIG=local/server_local_dev.yaml
```

启动机器人前必须准备并启用相应 QQ/CQHTTP 配置；禁用状态不代表已有可用机器人。Gateway 提供代理和静态站点服务，不会代替 API 进程或生成前端文件。监听地址、前端目录和证书配置见[配置说明](configuration.md)。

## 实际启动链

当前保留 `api`、`admin`、`robot`、`gateway` 四个正式命令，以及暂留的 `init`、`sync_wca` 运维命令。旧迁移程序与 `add-group`、`update-qq-group` 临时群组命令已移除；群组管理使用现有管理 API。

1. `main.go` 调用 `cmd.NewRootCmd().Execute()`。导入包的 `init()` 会在 `main()` 前执行。
2. `cmd/root.go` 的 `PersistentPreRunE` 调用 `svc.NewAPISvc(config, runJob, syncWca, runScramble)`。`--job/-j`、`--sync-wca/-w` 默认均为 `false`；打乱组件的内部开关当前为 `true`，没有开放对应命令行标志。
3. `src/internel/svc/svc.go` 加载 YAML、创建内存缓存，通过同包 `database.go` 的 `newDB` 连接业务 SQLite/MySQL，然后构造打乱组件。
4. 若 `global.algTrainersPath` 非空，调用 `algs.Init`。它先加载该路径相邻的 `Algs/catalog.json` 及其中列出的公式文件，失败会阻止服务初始化；再尝试加载未被替换的 Trainer。不是启动时自动下载资源。
5. `wca.NewWCA` 创建独立 WCA 数据访问对象，先尝试读取 `global.wcaDB.dbPath/wca.txt` 并连接其中指定的 MySQL 库；`-w` 开启同步循环，否则仍有定期重新读取本地指针的循环。
6. `convenient.NewConvenient` 调用同包 `migrate.go` 的迁移逻辑，尝试对业务模型 `AutoMigrate`（包括 User 和 OAuthState），随后通过 `jobs.go` 组装任务。`-j` 才运行任务循环；迁移不受 `-j` 控制，且当前迁移错误被忽略，进程继续不等于迁移成功。
7. 根命令设置滚动日志，才进入子命令 `RunE`。`api` 调用 `src/api.NewAPI`，初始化 JWT、鉴权和 `/v3/cube-api` 路由，最后监听 `api.host:api.port`；`gateway`、`robot` 分别进入对应实现。

因此 Gateway、Robot 等服务子命令也会走共享服务初始化，可能连接和迁移业务数据库。不要把它们理解成只读取 YAML 的轻量工具。以上文件拆分仍在原包内，不增加新的启动步骤或改变行为。

## 可选依赖与运行边界

### 打乱与图片

`global.scramble.type: tnoodle` 调用外部 HTTP 服务，代码位于 `src/internel/scramble/tnoodle_scramble.go`，使用 `/frontend/data/events`、`/frontend/puzzle/{id}/colors`、`/frontend/puzzle/{id}/scramble` 等接口。需确认实际服务版本支持这些接口；本仓库的 Go 命令不会启动 tNoodle。图片的 `2mf8` 分支在 `scramble_images.go` 使用固定的外部服务 URL；虽然配置的 `global.scramble.scramble` 被保存在对象里，当前绘图代码并不读取它，不能通过该字段切换图片服务。

macOS 由 `rust_scramble_darwin.go` 提供 `rust_twisty` 占位实现，返回 `[macOS-disabled]`，并非真实打乱。其他平台编译 `rust_scramble.go`，经 CGO 链接静态库；即使运行时选择 `tnoodle`，根入口构建仍会编译这个包。

### WCA、任务与联网

业务库可使用 SQLite，但当前 `wca.NewWCA` / `updateDb` 的查询路径使用独立 MySQL 数据库；SQLite 导出功能不代表可用业务 SQLite 代替 WCA 查询库。没有准备 WCA 数据时不能据此宣称相关接口可用。

`-w` 会进入 WCA 远程下载、解压、MySQL 导入、索引与清理流程，涉及数据库删除/创建和本地文件写入，需隔离数据库、足够磁盘及 `mysql` 等外部命令条件。`-j` 可运行成绩/赛事更新；启用静态站点自动更新时，还可能执行 `git pull` 和配置的构建命令。两者都应显式按需启用。

即使不传 `-j` / `-w`，导入 `src/internel/wca_api` 仍会在 `wcaSeniors.go` 的 `init()` 启动协程：尝试读取缓存，缺失或过期时请求远端并写入 `wca_seniors_cache.json`，之后定时更新。共享服务的导入链可到达该包，关闭上述标志不是离线保证。

邮件、WCA OAuth、QQ/CQHTTP 机器人分别需要 SMTP、OAuth 应用和机器人/协议服务配置。私人凭据留在被忽略的本地配置中，不要粘贴到日志、Issue 或 PR。

## 按范围验证

以下是供开发者选择的命令，不表示已在任意环境验证通过。先检查目标包当前代码及传递导入，再执行。

### 基础配置检查

```bash
make config-test
# 直接执行配置包测试
go test -mod=mod ./src/configs
```

该范围只导入配置包及其依赖，不加载 `svc` 或整套 API 服务；当前测试检查公开示例字段、兼容键和加载错误，并在测试临时目录写入 YAML。它不会验证数据库、证书、SMTP、tNoodle 或 WCA 服务是否可达，也不是检查 `CONFIG` 指定文件的完整启动演练。

### 群比赛逻辑

```bash
go test -mod=mod ./src/internel/groupcomp
```

已按源码核对：该包及测试使用模型、工具和 GORM，不导入 `svc`、`convenient`、打乱或 `wca_api`。`results_test.go` 的夹具使用 `sqlite.Open(":memory:")` 并在内存库迁移模型；`contract_test.go` 通过 AST/文件读取检查路由，不导入 HTTP Handler。可独立选择此包做群比赛回归，但仍需 CGO/C 编译器；依赖下载或 Go 构建缓存写入也需考虑。这个结论不能推广到父目录或 `./...`。

### 构建与全量测试的区别

```bash
make build
# 同一根入口的兼容目标
make build_root
```

`go build` 编译和链接，不执行应用包的 `init()` / `main()`，但仍需要依赖、可用的原生库和编译资源，例如 `src/internel/ttf/HuaWenHeiTi.ttf` 的 `go:embed` 文件。构建成功不证明数据库迁移、外部服务、资源运行路径或业务接口正确；运行二进制时才会执行初始化。

不要默认执行全量 `go test ./...`，也不要将 `go test ./... -run '^$'` 当成纯编译检查。Go 测试二进制在筛选测试函数前仍会执行导入包的 `init()`，可能联网、写缓存或启动后台协程。当前部分 WCA/统计测试使用真实数据库、同步或生成文件，不适合作为新贡献者的默认检查。

执行此类测试前必须逐项确认包和测试名、网络目标、隔离数据库、工作目录、写入路径及清理条件，显式选择范围。`-run` 仅筛选测试函数，不能隔离初始化。不要对生产数据库运行这些测试。

文档修改可做只读路径/命令核对与 `git diff --check`，并如实报告未运行应用或测试。详细提交要求见[贡献指南](../CONTRIBUTING.md)。

## 依赖与数据维护

`vendor/`、`local/`、原始快照 `test/` 被 Git 忽略，公开克隆不会因此自带开发者的本地内容。已有 vendor 时注意 Go 的 vendor 选择；需要按 `go.mod` 解析可用 `-mod=mod`，不要因文档改动重新整理依赖。

运行时公式数据在 `build/Algs/` 和 `build/Alg-Trainers/`。`scripts/import_local_algs.py` 的导入和 `--check` 都依赖未跟踪的 `test/2x2db/`、`test/algs/` 原始快照，公开克隆不能直接完成这种来源比对。脚本依赖、参数与写入行为见[脚本说明](../scripts/README.md)。
