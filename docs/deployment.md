# 部署说明

本文说明仓库现有 systemd 文件和 Makefile 的真实行为，不修改其生产路径。先阅读 [配置参考](configuration.md) 和 [etc 目录说明](../etc/README.md)。公开 YAML 是本地开发示例，部署前必须制作私人配置并准备功能依赖。

## 本地开发与构建

仓库 `go.mod` 声明 Go 1.25.0，需兼容的 Go 工具链。业务 SQLite 驱动需要 CGO / C 工具链；非 macOS 的打乱包还链接 `librust_scramble`，即使运行时选择 tNoodle 也不能跳过构建链接依赖。macOS 的 Rust 打乱实现为占位，不能据此认为完整打乱功能可用。

在仓库根目录运行：

```sh
make init-config
make config-test
make build
```

`CONFIG` 默认 `./local/server_local_dev.yaml`，可对需要配置的目标指定其他文件。`init-config` 只复制不存在的目标，已有文件和符号链接均保留。`config-test` 仅运行配置包测试，覆盖公开示例字段、加载错误及网关兼容优先级，不启动应用；它不验证外部服务、任意 `CONFIG` 文件或业务功能。构建可能需要下载 Go 依赖，不会安装或启动外部业务服务。

确认配置和依赖后，再按需使用 `make api`、`make gateway`、`make robot`；普通 `make` 仍等价于 `make api`，会启动程序。运行程序可能迁移数据库并触发后台初始化，不应当作无副作用的配置检查。邮件、OAuth、WCA 数据、tNoodle 和前端产物均需另行准备。机器人默认禁用，WeChat 尚未接通。

## systemd 固定约定

| 项目 | 三个 service 当前值 |
| --- | --- |
| 可执行文件 | `/usr/local/bin/cubing-pro` |
| 配置文件 | `/etc/cubing_pro/config.yaml` |
| 工作目录 | `/usr/local/bin` |
| 子命令 | API: `api`；网关: `gateway`；机器人: `robot` |
| 重启 | `Restart=always`，间隔 1 秒 |
| 运行用户 | 未设置 `User` / `Group`，系统服务默认 root |
| 任务参数 | 均未带 `-j` 或 `-w` |

修改 Makefile 的 `CONFIG` 不会改这些 service 的 `ExecStart`。service 中的 `After=rc-local.service` 也不代表已准备网络、MySQL、tNoodle 或其他依赖。只启用需要的服务，启动 gateway 不会自动启动 API 或 tNoodle。

## 首次安装

以下是在 Linux/systemd 主机上由部署人员执行的步骤；需 sudo 或等效权限，不是 `make build` 的副作用。已有配置应先备份并逐项审查，不能用公开示例覆盖生产值。

1. 构建 `./cubing-pro`，确认目标系统的 CGO / Rust 库等运行依赖。安装二进制：

```sh
sudo install -m 0755 cubing-pro /usr/local/bin/cubing-pro
```

2. 创建配置目录，首次复制示例；`cp -n` 保留已存在的目标，不合并字段：

```sh
sudo install -d -m 0750 /etc/cubing_pro
sudo cp -n etc/server.yaml /etc/cubing_pro/config.yaml
sudo chmod 0600 /etc/cubing_pro/config.yaml
```

3. 在机器上编辑 `/etc/cubing_pro/config.yaml`，改为真实部署参数。所有可配置的数据、日志、资源、公式及证书路径使用绝对路径，例如业务 SQLite `/var/lib/cubing-pro/cubingPro.db`、API 上传资源 `/var/lib/cubing-pro/static`、日志 `/var/log/cubing-pro`、前端产物 `/srv/cubing-pro/frontend/out`。自行创建目录、放置产物并授予实际运行账号权限；后端不会复制前端文件或下载公式库。不要将这些数据放到二进制目录或配置目录。

4. 手动安装所需的 service 文件并重新加载 systemd（以下列出全部三个，按部署范围选择）：

```sh
sudo install -m 0644 etc/cubing_pro_api.service /etc/systemd/system/cubing_pro_api.service
sudo install -m 0644 etc/cubing_pro_gw.service /etc/systemd/system/cubing_pro_gw.service
sudo install -m 0644 etc/cubing_pro_robot.service /etc/systemd/system/cubing_pro_robot.service
sudo systemctl daemon-reload
```

5. 按下文配置运行账号、目录权限和任务归属；确认后才启用需要的服务。例如仅启用 API 和网关：

```sh
sudo systemctl enable --now cubing_pro_api.service cubing_pro_gw.service
sudo systemctl status cubing_pro_api.service cubing_pro_gw.service
sudo journalctl -u cubing_pro_api.service -u cubing_pro_gw.service -n 100 --no-pager
```

机器人有可用实例和平台凭据时再单独启用 `cubing_pro_robot.service`。不要把含凭据或敏感请求的配置 / 日志原文提交到 issue 或仓库。

## 权限与工作目录

现有 service 默认 root 运行。若改为专用账号，先创建账号和所需目录，通过 `sudo systemctl edit <service>` 配置本机 override 的 `User` / `Group`，再授予配置读取、父目录遍历、日志 / 数据目录写入权限；上面的配置 `0600` / 目录 `0750` 也要按实际账号调整属主或组权限。证书私钥只需可读，不应公开。监听 80/443 等特权端口需要相应权限或能力，也可在前置反代后使用非特权端口。网关自身绑定所有网卡，应配置防火墙及入口访问策略。

所有相对文件路径按 cwd 解析。现有 service 的 `/usr/local/bin` 会使 `./cubingPro.db`、`./logs`、`./local/static` 等落到二进制目录下，因此不能直接部署本地示例。Seniors 缓存固定使用 cwd 下 `wca_seniors_cache.json`，QQ webhook 另有 cwd 下的 `logs/`，它们不随 `log.path` 改动。应通过本机 override 为需要的服务设置可写的专用 `WorkingDirectory`，同时将可配置文件路径改为绝对路径；只修改绝对路径不能迁移这些硬编码相对路径。变更工作目录前确认旧数据位置，避免无意新建另一个空 SQLite 库。

需要公式功能时，必须成对部署同级目录：

```text
/srv/cubing-pro/formulas/
  Alg-Trainers/
  Algs/
    catalog.json
    ...catalog 引用的数据文件
```

设置 `global.algTrainersPath` 为 `/srv/cubing-pro/formulas/Alg-Trainers`。初始化先读同级 `Algs/catalog.json` 并校验数据，缺失 / 无效将使共享服务初始化失败。只放 `Alg-Trainers` 不够。

## 任务与网络依赖

- 三个 service 共享服务初始化，均会打开业务库并尝试迁移，独立进程启动不等于各自只使用一组配置。
- 常规任务由 `-j` 开启，`job.unRunJobs` 按任务名排除；任务会在启动时立即执行一次。给共享数据或同一前端仓库只指定一个执行实例。静态站点自动更新需 `staticSites[].autoUpdate: true` 和 `-j`，会执行 `git pull` 和构建命令，运行账号须有仓库写权限及工具链。
- WCA 同步由独立的 `-w` 开启，`-j` 和 `unRunJobs` 不能代替或关闭它。只允许一个进程对同一 WCA 数据目录 / MySQL 实例执行同步；同步会建库、删旧库，需独立授权及可写暂存目录。不带 `-w` 的进程仍会读 `wca.txt` 并每小时尝试刷新数据库连接。
- 若要让 API 承担任务，使用本机 `systemctl edit` 清空并重设该服务的 `ExecStart`，保留二进制和配置的绝对路径，按需加 `-j` 和 / 或 `-w`；另外两个服务不要重复添加。仓库中的 unit 不会自动替你指定任务实例。
- Seniors 包级初始化会读缓存、在需要时联网获取数据，并每 6 小时更新，不受 `-j`、`-w`、机器人开关或空 WCA 配置控制。关闭任务开关不代表完整离线。
- 打乱请求依赖独立的 tNoodle 兼容服务；绘图、OAuth、邮件及数据抓取还各有网络依赖。生产应验证需要的具体流程，不能以二进制构建成功或 HTTP 监听成功作为全部业务可用的证明。

示例中 `httpsPort: 0` 且证书和私钥为空，确实关闭网关 TLS。启用内置 TLS 需同时设置正数 `httpsPort`、`pem`、`privateKey`；此时 HTTP 通常转向 HTTPS。若由前置代理终止 TLS，请按实际公开 URL 设置 `baseHost` 和 OAuth 回调，审查转发头及网关暴露范围。Next 静态多页导出保持 `spa: false`，保留动态占位规则且部署对应 HTML；遗留 Umi 配置及优先级见配置参考。

## 更新与回滚

`make install` 保持旧行为：构建二进制，停止三个服务，复制到 `/usr/local/bin/cubing-pro`，再重启三个服务。它需要构建及系统管理权限，不会安装配置 / unit、执行 `daemon-reload`、备份数据或准备外部资源；`CONFIG` 不参与 install。即使你只计划运行部分服务，此目标仍操作全部三个服务，部署前必须理解这一点。

需要只更新部分服务时，可手动备份现有二进制、私人配置与数据库，构建后停止选定服务、安装二进制并重启；若更新了本机 unit / override，先 `daemon-reload`。启动时可能自动迁移业务表，回滚二进制前须评估数据兼容性并按备份计划恢复，不能承诺只换回二进制就能回滚数据库。

公开示例调整不自动迁移现有私人 / 生产配置。参考配置文档的迁移表逐项更新：无效旧字段需移除或映射到实际消费者，旧键 `apiGateway` 的兼容不是字段合并。贡献部署改动时同步文档与非敏感示例，避免改写某台机器的真实配置。
