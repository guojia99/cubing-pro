# 公式库目录与公式论坛

本文档说明公式库迁入 MySQL 与公式论坛功能的启动并发控制、表结构约定及日常维护方式，详细接口参数以代码为准（[handlers.go](../src/api/app/formulaforum/handlers.go)）。

## 多进程并发启动锁与迁移错误传播

### 为什么需要启动锁

`Makefile` 的 `install` 目标和生产 systemd 运维会在同一时刻重启三个共用同一数据库配置的独立进程：`cubing_pro_api.service`、`cubing_pro_gw.service` 和 `cubing_pro_robot.service`；而 `cmd/root.go` 通过 `PersistentPreRunE` 让所有子命令在启动阶段统一调用 `svc.NewAPISvc`。

GORM `AutoMigrate` 会先探测表是否存在再发送不带 `IF NOT EXISTS` 的 DDL 建表语句，首次发布或新增表结构时多进程同时启动会抢建表并导致部分进程启动失败。

### 实现机制（MySQL `GET_LOCK`）

- 核心封装见 [migration_lock.go](../src/internel/database/migration_lock.go) 中的 `database.WithMigrationLock`：
  - 使用 `db.Connection` 固定单一底层物理连接，并通过 `Session(&gorm.Session{NewDB: true})` 在同一会话内执行 `SELECT GET_LOCK(?, 120)` 与 `defer SELECT RELEASE_LOCK(?)`。
  - 锁名称分别为 `cubing_pro_formula_init`（保护 `algs.InitCatalog` 的 6 张目录表 `AutoMigrate` 与首次事务导入）和 `cubing_pro_forum_migrate`（保护 `formulaforum.Migrate` 的 3 张论坛表 `AutoMigrate`）。
  - 零额外组件、零新增锁表；非 MySQL 方言（如单元测试使用的 SQLite）直接执行迁移回调不加锁。
- 错误严格向上传播（fail-fast）：
  - `algs.InitCatalog` 与 `formulaforum.Migrate` 均在 `svc.NewAPISvc` 中执行并检查返回错误（不再放在忽略错误的 `convenient.NewConvenient` 中），任意一步建表、获取/释放锁或导入校验失败均直接中止启动，避免进程带病运行到访问阶段才报错。
  - 若需改为由外部发布脚本单进程完成迁移，只需要去掉上述两处的 `WithMigrationLock` 包裹并移除 `migration_lock.go`。

## 公式目录入库、稳定 ID 与维护方式

- 数据库表（[catalog.go](../src/internel/algs/catalog.go)）：`formula_catalog_imports`、`formula_classes`、`formula_sets`、`formula_groups`、`formula_cases`、`formula_solutions`。
- **16 位小写 hex 稳定 ID**：首次导入时由层级键与动作串计算 SHA-256 前 8 字节小写 hex（16 字符；数据库列 `size:32` / `varchar(32)`，接口 URI 校验 `len=16,hexadecimal,lowercase`）。
- **首版口径**：
  - 初次上线保持原有展示顺序与动作串原文，确保客户端既有的本地首选下标记录与分享深链持续有效。
  - `333oh/ZBLL` 中 `ADN11` 等相同动作串的重复槽位按两行独立解法导入，各自拥有独立 `id` 与讨论；空槽位在数据库内保留占位行以维持下标，对外接口的 `solutionIds` 对应位置返回 `null`。
- **后续在 MySQL 中的维护方式**：
  - 调整公式顺序或勘误公式文本时，直接更新 `formula_solutions.position` 或 `formula_solutions.moves` 并**保持 `id` 不变**，随后清理或重启 API 内存缓存（`svc.Cache`）即可生效。

## 公式论坛表结构与接口概览

- 数据模型（[forum.go](../src/internel/database/model/formulaforum/forum.go)）：
  - `formula_discussions`：`group_id` / `solution_id`（`varchar(32)` 唯一索引），数据库 `CHECK` 约束 `chk_formula_discussion_target` 强制二者二选一非空。
  - `formula_comments`：支持 `parent_id` 一层或多层回复、`deleted_at` 软删除（存在活跃直接子回复时前端保留“已删占位”）。
  - `formula_reply_notifications`：回复非本人评论时在同一事务内写入通知，删除对应回复时一并清理未读/关联通知。
- HTTP 接口（均挂在 `/v3/cube-api` 下，分页默认 `size=20`，最大 `100`）：

| 权限 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| 公开 | GET | `/public/formula-forum/counts` | 按 `cube`、`class` 返回各组与解法的活跃评论数 |
| 公开 | GET | `/public/formula-forum/targets/:kind/:targetId/comments` | 读取分组/解法讨论，支持 `commentId` 居中定位与越界 `page` 收敛 |
| 公开 | GET | `/public/formula-forum/recent` | 最近活跃讨论列表，返回 `lastCommentId` 便于直达最新评论 |
| 登录用户 | POST | `/user/formula-forum/targets/:kind/:targetId/comments` | 发评/回复（1–400 字校验、屏蔽词返回 422 `keyword_blocked`） |
| 登录用户 | GET | `/user/formula-forum/notifications` | 「回复我的」通知列表与未读数 |
| 登录用户 | PUT | `/user/formula-forum/notifications/:id/read` | 标记通知已读 |
| 管理员 | GET | `/admin/formula-forum/comments` | 按时间倒序查看活跃发言 |
| 管理员 | DELETE | `/admin/formula-forum/comments/:id` | 软删发言并清理通知及失去活跃子回复的已删祖先 |
