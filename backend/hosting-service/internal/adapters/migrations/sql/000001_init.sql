-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TYPE edition_type AS ENUM ('bedrock', 'java'); 

CREATE TABLE IF NOT EXISTS servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,

    name VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    port INTEGER NOT NULL CHECK (port BETWEEN 1 AND 65535),
    edition edition_type NOT NULL,

    CONSTRAINT servers_endpoint_unique
        UNIQUE (address, port, edition),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS servers;
DROP TYPE IF EXISTS edition_type;