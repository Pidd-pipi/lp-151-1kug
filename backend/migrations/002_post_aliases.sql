-- 002_post_aliases.sql：每帖树洞化名表。
-- 实际表结构由 GORM AutoMigrate 自动创建（见 internal/model/model_post_alias.go），
-- 此处保留参考定义。历史帖子与评论无需手工回填：升级后首次读取列表/详情/评论时，
-- 后端会按 (post_id, identity_id) 即时生成并落库，同一帖内保持稳定。
CREATE TABLE IF NOT EXISTS post_aliases (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    post_id     BIGINT UNSIGNED NOT NULL,
    identity_id BIGINT UNSIGNED NOT NULL,
    nickname    VARCHAR(64)  NOT NULL,
    avatar      VARCHAR(255) NULL,
    created_at  DATETIME(3)  NULL,
    UNIQUE KEY uk_post_alias_post_identity (post_id, identity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
