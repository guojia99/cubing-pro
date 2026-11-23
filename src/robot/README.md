# 机器人

当前 robot 子命令接入 QQ 官方机器人适配器和 CQHTTP 适配器；两者协议、凭据和接收方式不同。WeChat 只有配置声明，尚未接入实例创建流程。公开的 [`etc/server.yaml`](../../etc/server.yaml) 将所有实例设为 `enable: false`，不会自动安装、登录或启动外部机器人平台。

先阅读 [配置参考](../../docs/configuration.md) 和 [部署说明](../../docs/deployment.md)，在仓库根目录创建私人副本：

```sh
make init-config
make config-test
# 编辑私人副本并准备所需服务后，再按需执行：
make robot CONFIG=./local/server_local_dev.yaml
```

复制操作不会覆盖已有配置。全部禁用时 robot 仍会进入等待，并经历共享服务初始化；它不是无副作用的配置检查。业务数据库迁移、WCA 指针刷新及 Seniors 的包级缓存 / 联网行为并不由机器人开关全部控制，不能宣称离线运行。

## QQ

`robot.QQBot[]` 的准确 YAML 字段如下，示例的空值 / 零值只是占位，不能直接启用：

```yaml
robot:
  QQBot:
    - enable: false
      qq: 0
      app_id: 0
      token: ""
      app_secret: ""
      is_sandbox: false
      wss_addr: ""
      is_open: false
      server:
        port: 0
        certFile: ""
        certKey: ""
        isOpen: false
```

需要有效的 QQ 平台应用、权限与凭据、可达回调及匹配的 WebSocket 地址。启用后实现会同时启动内置 webhook 服务、用 `app_id` / `app_secret` 获取 access token、连接 `wss_addr`；`token` 被传入 webhook 设置，不是当前 access token 获取入口。没有 `${VAR}` 插值支持，应在受保护的私人配置中填写实际值。

`enable` 才是实例开关。`is_open` 在两种 WebSocket 连接方式之间选择；`server.isOpen` 在 `/websocket` 与 `/wss/qqbot` 等 webhook 路由模式之间选择，两者设为 `false` 也都会连接 / 监听。`server.port: 0` 会使用临时端口，不能直接搭配固定 `wss_addr`；须显式配置端口和匹配的 ws/wss 地址。`server.certFile` / `certKey` 均非空才启用此 webhook 的 TLS，与 gateway TLS 独立。webhook 监听失败还可能尝试其他端口，需检查实现及日志。

`is_sandbox` 影响客户端 OpenAPI，但 webhook 应用的沙箱标志当前硬编码 `false`，不能当成完整沙箱隔离。内置 webhook 使用全局设置，多实例不能仅凭多个 YAML 条目就假定相互独立。当前消息桥接主要处理群 @ 消息；旧示例中的 `group`、`appID`、`groupList` 不被配置结构读取，也没有对应的群白名单开关。

## CQHTTP

`robot.CQHttpBot[]` 使用 `enable`、`prefix`、`address`、`post`。`address` 是外部兼容 CQHTTP 服务的 HTTP API 根地址，发送群消息追加 `/send_group_msg`；`post` 是本进程监听 `127.0.0.1:<post>` 的整数反向 HTTP 端口，拼写不是 `port`。外部平台需自行配置，把群消息事件回送到该端口；后端不会代为部署或登录 go-cqhttp。

`prefix` 虽声明为命令前缀并由适配器的 `Prefix()` 返回，但当前消息分发没有调用该方法，不能依赖它过滤或去掉前缀。图片使用 `file://` 路径发送，外部进程 / 容器必须能读取相应文件。不要将 QQ 官方机器人的凭据和 WebSocket 配置套用到此适配器。

## 功能依赖与贡献

赛事、成绩、绑定等插件需要业务库和相应数据；打乱及 PK 计时功能需要可用打乱服务，图片、WCA 查询及 Seniors 等功能还可能请求外部服务。`robot.personValPath` 虽有素材插件消费者，但 `NewPlugins` 当前未注册人物素材插件，单独设置路径不会启用。`robot.WeChatBot[].enable` 当前也不会启动微信机器人。

配置相对文件路径一律按进程 cwd 解析，systemd 的 cwd 和本地仓库不同。凭据、日志及私人素材不要提交。新增机器人字段或功能时，请同时核对配置标签、实例创建和实际消费者，补充非敏感禁用示例、依赖说明及有针对性的测试；不要仅添加 YAML 就宣称功能可用。
