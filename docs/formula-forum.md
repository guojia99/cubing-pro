# 公式库目录与公式论坛

公式库和论坛共用服务数据库（MySQL 或 SQLite）。论坛接口实现见 [handlers.go](../src/api/app/formulaforum/handlers.go)。

## 安装与首次导入

`make install` 只安装二进制并重启服务，不复制公式文件，也不执行导入。`api`、`gateway` 和 `robot` 启动时只迁移目录表和论坛表，不读取 `Alg-Trainers` 或 `Algs/catalog.json`。新库可以先启动；目录列表为空时不写缓存，首次导入提交后下一次请求即可看到数据。

在有公式源文件的机器上，使用与服务相同的数据库配置单独导入。对 SQLite，跨工作目录运行时，配置中的 DSN 也要指向同一个数据库文件（建议用绝对路径）。例如：

```sh
cubing-pro init formulas --config /etc/cubing-pro/server.yaml --path /srv/cubing/build/Alg-Trainers
```

`--path`（`-p`）指向 Trainer 目录，同级须有 `Algs/catalog.json` 及清单引用的数据文件。不传路径时读取 `global.algTrainersPath`。开发环境可在仓库根目录执行 `make init_formulas FORMULAS_CONFIG=/path/to/server.yaml`。导入和完成标记在同一事务内提交；导入失败可修正后重试，成功后重跑不会覆盖库内修改。

## 并发启动与字符集

- [迁移锁](../src/internel/database/migration_lock.go) 保护目录建表、首次导入和论坛建表。MySQL 在同一物理连接上获取和释放命名锁：`cubing_pro_formula_init` 用于目录，`cubing_pro_forum_migrate` 用于论坛。获取、执行和释放失败均向调用方返回错误。
- 文件型 SQLite 按数据库绝对路径共用进程内互斥锁，并锁定同目录的 `<数据库文件>.migrate.lock`（Linux/macOS 使用 `flock`）。服务账号须能在该目录创建锁文件；运行期间不要删除或替换锁文件。内存库不创建锁文件。
- MySQL 连接使用 `charset=utf8mb4`。新建的目录表和论坛表使用 InnoDB、`utf8mb4_unicode_ci`，可保存含 emoji 的公式说明、评论及回复。若已有表使用三字节 `utf8`，需另行转换相关表和文本列；更换 DSN 不会转换既有数据表。

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
  - `formula_reply_notifications`：回复非本人评论时在同一事务内写入通知，删除对应回复时一并清理关联通知。
- **目录目标失效**：分组或解法被删除、解法动作被清空时，公开计数与最近讨论隐藏该目标，读帖和发帖返回 404。管理列表和回复通知保留原目标的 `kind`、`id`，并返回 `target.unavailable: true`；列表总数和未读数照常统计，仍可删评或标记通知已读。
- HTTP 接口（均挂在 `/v3/cube-api` 下，分页默认 `size=20`，最大 `100`）：

| 权限 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| 公开 | GET | `/public/formula-forum/counts` | 按 `cube`、`class` 返回各组与解法的活跃评论数 |
| 公开 | GET | `/public/formula-forum/targets/:kind/:targetId/comments` | 读取分组/解法讨论，支持 `commentId` 居中定位与越界 `page` 收敛。指定评论已删除且不再可见时，仍按收敛后的页码打开讨论 |
| 公开 | GET | `/public/formula-forum/recent` | 最近活跃讨论列表，返回 `lastCommentId` 便于直达最新评论 |
| 登录用户 | POST | `/user/formula-forum/targets/:kind/:targetId/comments` | 发评/回复（1–400 字，屏蔽词返回 422 `keyword_blocked`）。每个 API 进程按用户 ID 限制所有目标合计每滚动 60 秒 10 次请求，超出返回 429；同 IP 的其他用户不共用额度 |
| 登录用户 | GET | `/user/formula-forum/notifications` | 「回复我的」通知列表与未读数 |
| 登录用户 | PUT | `/user/formula-forum/notifications/:id/read` | 标记通知已读 |
| 管理员 | GET | `/admin/formula-forum/comments` | 按时间倒序查看活跃发言 |
| 管理员 | DELETE | `/admin/formula-forum/comments/:id` | 软删发言并清理通知及失去活跃子回复的已删祖先 |
