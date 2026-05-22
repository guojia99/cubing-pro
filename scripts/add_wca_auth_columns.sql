-- WCA OAuth 登录相关字段迁移
-- 执行: mysql -u user -p database_name < add_wca_auth_columns.sql
-- 或手动在 MySQL 客户端中执行

ALTER TABLE `users` ADD COLUMN `wca_login_at` DATETIME NULL AFTER `wca_id`;
ALTER TABLE `users` ADD COLUMN `wca_access_token` VARCHAR(512) NULL AFTER `wca_login_at`;
ALTER TABLE `users` ADD COLUMN `wca_token_expires_at` DATETIME NULL AFTER `wca_access_token`;
