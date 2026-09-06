-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE profile_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS profiles(
    user_id UUID PRIMARY KEY,
    username VARCHAR UNIQUE,
    bio VARCHAR(4096),
    avatar_url VARCHAR,
    role profile_role NOT NULL DEFAULT 'user',
    is_banned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blocks(
    user_id UUID NOT NULL,
    target_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (user_id, target_id),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (target_id) REFERENCES profiles(user_id)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    user_id    UUID  NOT NULL,
    target_user_id  UUID,
    target_topic_id BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (target_user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (target_topic_id) REFERENCES topics(id),

    CHECK(
        (target_topic_id IS NOT NULL AND target_user_id IS NULL)
        OR
        (target_topic_id IS NULL AND target_user_id IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS topics (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR UNIQUE NOT NULL,
    is_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS posts(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id UUID,
    user_id UUID NOT NULL,
    image_url TEXT,
    description VARCHAR,
    views INTEGER DEFAULT 0,

    is_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL,
    user_id UUID NOT NULL,
    parent_id UUID,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (post_id) REFERENCES posts(id),
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES profiles(user_id)
);

CREATE TABLE post_likes (
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (user_id, post_id),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id)
);

CREATE TABLE comment_likes (
    user_id UUID NOT NULL,
    comment_id UUID NOT NULL, 
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (user_id, comment_id),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id)
);

CREATE TABLE IF NOT EXISTS post_reports (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    cause VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id),

    UNIQUE (user_id, post_id)
);

CREATE TABLE IF NOT EXISTS comment_reports (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    comment_id UUID NOT NULL, 
    cause VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id),

    UNIQUE (user_id, comment_id)
);

CREATE TABLE IF NOT EXISTS profile_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    target_id UUID NOT NULL, 
    cause VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES profiles(user_id),
    FOREIGN KEY (target_id) REFERENCES profiles(user_id),

    UNIQUE (user_id, target_id)
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS reports;
