-- +goose NO TRANSACTION
-- +goose Up
CREATE UNIQUE INDEX CONCURRENTLY idx_profile_username ON profiles(username);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX posts_description_trgm ON posts USING gin (description gin_trgm_ops) WHERE is_delete = false;
CREATE INDEX streams_title_trgm ON streams USING gin (title gin_trgm_ops);
CREATE INDEX profiles_username_trgm ON profiles USING gin (username gin_trgm_ops);
CREATE INDEX profiles_username_prefix ON profiles (lower(username) text_pattern_ops);
CREATE INDEX streams_title_prefix ON streams (lower(title) text_pattern_ops);
-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_profile_username;
DROP INDEX IF EXISTS streams_title_prefix;
DROP INDEX IF EXISTS profiles_username_prefix;
DROP INDEX IF EXISTS profiles_username_trgm;
DROP INDEX IF EXISTS streams_title_trgm;
DROP INDEX IF EXISTS posts_description_trgm;