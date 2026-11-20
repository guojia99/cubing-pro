# WCA OAuth 实现索引

本目录使用项目现有 Gin 路由、数据库和 JWT 中间件实现 WCA 登录，无需另建示例项目或执行 `go mod init`。前端参数、配置与响应说明见 [WCA 登录前端对接手册](../../../../docs/WCA_LOGIN_FRONTEND_GUIDE.md)。

## 实现位置

| 文件 | 职责 |
| --- | --- |
| [wca.go](wca.go) | `WcaAuthInit` 发起授权、持久化 state；`WcaAuthCallback` 换取 WCA token、读取用户信息并生成本地 JWT；`WcaAuthMe` 返回当前鉴权用户 |
| [routes/auth.go](../../routes/auth.go) | 注册 `/v3/cube-api/auth/wca`、`/callback` 和需 JWT/选手权限的 `/me` 路由 |
| [middleware/jwt.go](../../middleware/jwt.go) | 与账号密码登录共用的 JWT 生成和解析逻辑 |
| [user/oauth_state.go](../../../internel/database/model/user/oauth_state.go) | 数据库中的 nonce、前端跳转地址与 state 过期时间 |
| [configs/global.go](../../../configs/global.go) | `global.wcaAuth2` 配置结构 |
| [convenient/migrate.go](../../../internel/convenient/migrate.go) | 服务初始化时尝试迁移用户及 OAuth state 表；旧 MySQL 库的条件补迁移见[脚本说明](../../../../scripts/README.md) |

## 对接要点

- 配置位于 **`global.wcaAuth2`**，不从旧示例中的 `WCA_CLIENT_ID` 等环境变量读取。凭据仅放本地/部署配置，不提交仓库。
- 回调使用 `redirectBase`，为空时回退到 `global.baseHost`，再拼接 `/v3/cube-api/auth/wca/callback`。
- 前端 `redirect` 允许列表优先使用非空 `redirectURLs`；只有列表为空时才使用 `frontendBase` 和不同的 `redirectBase`。空值或 `debug` 返回 JSON 调试结果。
- 正常发起授权使用 **307** 跳转；登录成功后使用 **302** 跳回前端并在查询参数中附带本地 JWT。
- state 记录保存在业务数据库，当前有效期为 10 分钟；回调消费后删除。查不到/过期 state 或换 token 失败时，当前代码会尝试重新发起授权。
- WCA 用户需有 WCA ID；本地按 `wca_id` 查找或创建用户。WCA access token 不作为前端 API 的 Bearer token，也不会通过用户模型 JSON 返回。

启动与测试可能涉及数据库和导入包初始化，请先阅读[开发指南](../../../../docs/development.md)。
