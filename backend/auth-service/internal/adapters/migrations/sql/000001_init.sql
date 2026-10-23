-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users(
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash VARCHAR,
    picture_url VARCHAR,
    role user_role NOT NULL DEFAULT 'user',
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_banned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth_accounts(
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_user_id VARCHAR NOT NULL,
    given_name VARCHAR(255),
    family_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(provider, provider_user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS users;
