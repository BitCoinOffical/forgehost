-- +goose NO TRANSACTION
-- +goose Up

CREATE INDEX idx_users_email_unverified ON users(id) WHERE email_verified = false;
CREATE INDEX idx_users_email_verified ON users(id) WHERE email_verified = true;
-- +goose Down
DROP INDEX IF EXISTS idx_users_email_unverified;
DROP INDEX IF EXISTS idx_users_email_verified;