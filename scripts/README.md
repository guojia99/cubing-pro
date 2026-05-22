# 数据库迁移

MySQL 用户需手动执行以下迁移。将 `database_name` 替换为实际数据库名。

## 修复 AutoMigrate 报错（若启动时报错 1062、1170 等，先执行）

```bash
mysql -u 用户名 -p database_name < fix_automigrate_errors.sql
```

## WCA OAuth 相关迁移（必须执行）

```bash
# 1. users 表新增 WCA 登录字段
mysql -u 用户名 -p database_name < add_wca_auth_columns.sql

# 2. 新建 oauth_states 表
mysql -u 用户名 -p database_name < add_oauth_states_table.sql
```

或依次执行：
```bash
mysql -u 用户名 -p database_name < add_wca_auth_columns.sql
mysql -u 用户名 -p database_name < add_oauth_states_table.sql
```

## 手动执行 SQL

若无法使用命令行，可在 MySQL 客户端中执行：

**add_wca_auth_columns.sql:**
```sql
ALTER TABLE `users` ADD COLUMN `wca_login_at` DATETIME NULL AFTER `wca_id`;
ALTER TABLE `users` ADD COLUMN `wca_access_token` VARCHAR(512) NULL AFTER `wca_login_at`;
ALTER TABLE `users` ADD COLUMN `wca_token_expires_at` DATETIME NULL AFTER `wca_access_token`;
```

**add_oauth_states_table.sql:** 见该文件内容。
