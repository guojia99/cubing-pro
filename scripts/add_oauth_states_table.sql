-- OAuth state 表，用于 WCA 登录流程的 state 持久化（防 CSRF，服务器重启后仍可校验）
-- 执行: mysql -u user -p database_name < add_oauth_states_table.sql

CREATE TABLE IF NOT EXISTS `oauth_states` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `nonce` varchar(64) NOT NULL,
  `redirect` varchar(512) DEFAULT NULL,
  `expires_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_oauth_states_nonce` (`nonce`),
  KEY `idx_oauth_states_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
