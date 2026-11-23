# 公式库目录与公式论坛

公式库和论坛共用服务 MySQL 数据库（8.0.16+）。论坛接口实现见 [handlers.go](../src/api/app/formulaforum/handlers.go)。

## API 启动与首次导入

共享服务通过 `convenient` 统一迁移目录、论坛和通知阅读表；只有 `api.NewAPI` 引导公式目录数据导入，`gateway` 和 `robot` 不读取公式源文件。

首次导入时，在 `api` 使用的 `server.yaml` 中设置 `global.algTrainersPath`，建议使用绝对路径：

```yaml
global:
  algTrainersPath: "/srv/cubing/build/Alg-Trainers"
```

- **路径留空或未配置**：API 跳过导入，直接使用已迁移的数据库，不读取公式源文件。新库的目录列表为空时不写缓存。
- **已配置路径、数据库尚未完成导入**：`api` 自动读取 Trainer 目录、同级的 `Algs/catalog.json` 及清单引用的数据文件。目录数据和 `formula_catalog_imports.complete` 在同一事务中提交。若源文件缺失，`api` 启动报错；修正路径或文件后重新启动即可。
- **数据库已完成导入**：重启时跳过读取源文件，保留库内修改。原导入目录不再需要。

## 公式目录入库、稳定 ID 与维护方式

- 数据库表（[catalog.go](../src/internel/database/model/formula/catalog.go)）：`formula_catalog_imports`、`formula_classes`、`formula_sets`、`formula_groups`、`formula_cases`、`formula_solutions`。
- **16 位小写 hex 稳定 ID**：首次导入时由层级键与动作串计算 SHA-256 前 8 字节小写 hex（16 字符；数据库列 `size:32` / `varchar(32)`，接口 URI 校验 `len=16,hexadecimal,lowercase`）。
- **首版口径**：
  - 初次上线保持原有展示顺序与动作串原文，确保客户端既有的本地首选下标记录与分享深链持续有效。
  - `333oh/ZBLL` 中 `ADN11` 等相同动作串的重复槽位按两行独立解法导入，各自拥有独立 `id` 与讨论；空槽位在数据库内保留占位行以维持下标，对外接口的 `solutionIds` 对应位置返回 `null`。
- **后续在 MySQL 中的维护方式**：
  - 调整公式顺序或勘误公式文本时，直接更新 `formula_solutions.position` 或 `formula_solutions.moves` 并**保持 `id` 不变**，随后清理或重启 API 内存缓存（`svc.Cache`）即可生效。

## 公式论坛表结构与接口概览

- 数据模型（[forum.go](../src/internel/database/model/formulaforum/forum.go)）：
  - `formula_discussions`：`set_id` / `group_id` / `solution_id`（`varchar(32)` 唯一索引），数据库 `CHECK` 约束 `chk_formula_discussion_scope` 强制三者恰有一个非空。
  - `formula_comments`：支持 `parent_id` 一层或多层回复、`deleted_at` 软删除（存在活跃直接子回复时前端保留“已删占位”）。
  - `formula_reply_notifications`：回复非本人评论时在同一事务内写入通知，删除对应回复时一并清理关联通知。
  - `site_announcement_reads`：按用户和公告 ID 保存统一通知中心的公告已读时间，公告编辑、撤回后重新发布沿用原 ID 和阅读状态。
- **目录目标失效**：集合、分组或解法被删除、解法动作被清空时，公开计数与最近讨论隐藏该目标，读帖和发帖返回 404。管理列表和回复通知保留原目标的 `kind`、`id`，并返回 `target.unavailable: true`；列表总数和未读数照常统计，仍可删评或标记通知已读。
- HTTP 接口（均挂在 `/v3/cube-api` 下，分页默认 `size=20`，最大 `100`）：

| 权限 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| 公开 | GET | `/public/formula-forum/counts` | 按 `cube`、`class` 返回各集合、分组与解法的活跃评论数 |
| 公开 | GET | `/public/formula-forum/targets/:kind/:targetId/comments` | 读取集合/分组/解法讨论，支持 `commentId` 居中定位与越界 `page` 收敛。指定评论已删除且不再可见时，仍按收敛后的页码打开讨论 |
| 公开 | GET | `/public/formula-forum/recent` | 最近活跃讨论列表，返回 `lastCommentId` 便于直达最新评论 |
| 登录用户 | POST | `/user/formula-forum/targets/:kind/:targetId/comments` | 发评/回复（1–400 字，屏蔽词返回 422 `keyword_blocked`）。每个 API 进程按用户 ID 限制所有目标合计每滚动 60 秒 10 次请求，超出返回 429；同 IP 的其他用户不共用额度 |
| 登录用户 | GET | `/user/formula-forum/discussions` | 当前用户参与过的公式讨论，按集合、分组或解法去重，返回发言总数和最新活动 |
| 登录用户 | GET | `/user/formula-forum/comments` | 当前用户未删除的公式评论和回复，返回所属公式及回复对象 |
| 登录用户 | GET | `/user/formula-forum/notifications` | 「回复我的」通知列表与未读数 |
| 登录用户 | PUT | `/user/formula-forum/notifications/:id/read` | 标记通知已读 |
| 登录用户 | GET | `/user/notifications?page&size&type` | 统一通知列表；`type` 为 `all`、`formula-reply` 或 `site-announcement`，返回当前分类 `total` 和全部通知未读数 `unread` |
| 登录用户 | PUT | `/user/notifications/:id/read` | 标记统一通知已读；ID 形如 `formula-reply:<id>` 或 `site-announcement:<公告ID>`，操作幂等并校验收件人或公告有效状态 |
| 管理员 | GET | `/admin/formula-forum/comments` | 按时间倒序查看活跃发言 |
| 管理员 | DELETE | `/admin/formula-forum/comments/:id` | 软删发言并清理通知及失去活跃子回复的已删祖先 |
