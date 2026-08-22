-- +goose NO TRANSACTION
-- +goose Up
CREATE UNIQUE INDEX CONCURRENTLY idx_profile_username ON profile(username);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_profile_username;