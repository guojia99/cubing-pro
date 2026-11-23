# 公开配置与部署文件

`server.yaml` 是可公开贡献的本地开发示例，不是完整业务环境，也不是生产配置。字段和行为见 [配置参考](../docs/configuration.md)，systemd 安装见 [部署说明](../docs/deployment.md)，生产上线前的核对与材料准备见[后端上线须知与材料清单](../docs/production-go-live.md)。

在仓库根目录执行：

```sh
make help
make init-config
make config-test
make build
```

`make init-config` 将本目录的示例复制到 `./local/server_local_dev.yaml`，目标存在时保留原文件（包括符号链接），也可以用 `CONFIG=/absolute/path/config.yaml` 指定另一份副本。它不会合并配置或迁移数据。修改公开示例不会覆盖已有私人配置或 `/etc/cubing_pro/config.yaml`；直接使用 `--config ./etc/server.yaml` 的进程则会在下次加载时读到新版示例。

示例保留 SQLite 的 `./cubingPro.db` 和 API 的 `localhost:20000`，网关改用 HTTP `8080`，显式关闭 TLS，前端静态产物示例位置为 `./local/frontend/out`，上传资源为 `./local/static`，所有机器人实例默认禁用。邮件、WCA OAuth、WCA 数据、公式库及打乱服务等需要单独准备，详见配置参考。所有相对文件路径都按进程工作目录解析。

## systemd 文件

三个 service 保持现有部署约定，不能当作本地启动配置使用：

| 文件 | 命令 | 二进制 / 配置 / 工作目录 |
| --- | --- | --- |
| `cubing_pro_api.service` | `api` | `/usr/local/bin/cubing-pro` / `/etc/cubing_pro/config.yaml` / `/usr/local/bin` |
| `cubing_pro_gw.service` | `gateway` | 同上 |
| `cubing_pro_robot.service` | `robot` | 同上 |

service 未设置 `User`/`Group`，系统服务默认以 root 运行。使用普通用户部署时，须在机器上的 override 中设置账号并授予配置读取、数据和日志目录写入权限；低于 1024 的监听端口还需要相应权限。部署配置的数据、日志、前端与证书路径应使用绝对路径，否则相对路径会落在 `/usr/local/bin` 下。Seniors 缓存等代码固定的相对路径也需要考虑该目录权限，详见部署说明。

`make install` 只保留原有的构建、停止三个服务、复制二进制、重启三个服务行为，需要相应系统权限。它**不会**复制配置、安装 service、执行 `daemon-reload`、创建数据目录或安装外部依赖；首次安装必须按部署说明手动完成。三个 service 默认都不带 `-j`/`-w`。定时任务和 WCA 同步应分别指定唯一实例承担，避免多个进程重复执行。公式数据启用时，`Alg-Trainers` 旁边必须有同级 `Algs` 目录及有效 `catalog.json`。

## 贡献与迁移

- 修改字段前同时核对 `src/configs/` 的 YAML 标签和实际消费者，补充配置参考及示例；仅声明字段不代表已经生效。
- 不提交 token、密码、私人域名或生产路径。`local/` 已被忽略，但自选的配置路径不一定被忽略，提交前检查暂存区。
- 新示例去掉了无效的 `global.xStaticPath`、`global.xFilePath`、`global.wca_db` 及 QQ 的 `group`、`appID`、`groupList`。`global.algPath` 同样没有消费者，公式库位置只由 `global.algTrainersPath` 控制。迁移时按配置参考逐项调整私人副本，不要直接覆盖现有生产配置。
- `make config-test` 仅运行配置包测试，不承诺完整配置校验或业务可用性；运行应用前确认外部依赖和后台行为。不要用全量测试代替本地配置检查，部分包可能联网或依赖外部资源。
