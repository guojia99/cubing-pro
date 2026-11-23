---
name: cubing-pro-backend
description: cubing-pro Go 后端的数据库与初始化约定。修改或审查业务模型、SQL、迁移、数据库测试和 API 初始化时使用。
---

# 后端数据库与初始化约定

- 业务数据库唯一指定为 **MySQL 8.0.16+**。配置只允许 `global.db.driver: mysql`。
- 不为 SQLite 添加任何保护性兼容或额外操作，包括数据库分支、表重建、索引或触发器保留、自增序列恢复、锁冲突重试、降级 SQL、驱动回退和专用迁移。
- 不为运行 SQLite 测试而改变 MySQL 业务逻辑、事务、锁或约束。业务数据库测试使用 MySQL，统一使用 `src/internel/testutil.OpenMySQL` 和 `CUBING_TEST_MYSQL_DSN`；缺少测试库时明确报告跳过，不回退到 SQLite。
- WCA 离线 SQLite 导出是用户明确保留的独立工具。保留其实现和必要依赖，不将它视为业务库支持 SQLite 的依据。
- 数据模型放在 `src/internel/database/model/`，建表及 MySQL 迁移由 `src/internel/convenient/` 统一引导；不在 model、API handler 或 cmd 中另设迁移入口。
- 公式目录数据导入由 API 子模块在 `NewAPI` 中引导，cmd 仅创建和启动 API。

变更配置、数据库支持或测试入口时，同步核对 `etc/server.yaml`、`docs/configuration.md`、`docs/development.md` 和对应功能文档。
