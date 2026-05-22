-- 修复 AutoMigrate 报错，需在启动应用前执行
-- 执行: mysql -u 用户名 -p 数据库名 < fix_automigrate_errors.sql

-- 1. 修复 users.login_id: 多个空字符串违反 UNIQUE，将空串改为 NULL（MySQL 允许多个 NULL）
UPDATE users SET login_id = NULL WHERE login_id = '';

-- 2. 修复 user_kvs.key: 若报错 BLOB/TEXT column 'key' used in key specification
-- 需将 key 改为 varchar 以支持索引（模型已加 size:191，若仍报错可手动执行）：
-- ALTER TABLE user_kvs DROP INDEX idx_user_key;
-- ALTER TABLE user_kvs MODIFY COLUMN `key` VARCHAR(191) NOT NULL;
-- ALTER TABLE user_kvs ADD UNIQUE INDEX idx_user_key (user_id, `key`);
