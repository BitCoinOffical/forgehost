-- +goose NO TRANSACTION
-- +goose Up

CREATE UNIQUE INDEX CONCURRENTLY idx_user_email ON users(email);
-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_user_email;