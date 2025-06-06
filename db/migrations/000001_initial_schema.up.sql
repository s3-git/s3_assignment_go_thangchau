-- Friend Management Database Schema
-- This schema supports the friend management system with friends, subscriptions, and blocks

-- Users table to store email addresses
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- Index for faster email lookups
CREATE INDEX idx_users_email ON users(email);

-- Friends table for bidirectional friend connections
-- Uses a constraint to ensure user1_id < user2_id to avoid duplicate entries
CREATE TABLE friends (
    id SERIAL PRIMARY KEY,
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_friends_user1 FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_friends_user2 FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Ensure user1_id is always less than user2_id to avoid duplicates
    CONSTRAINT chk_user_order CHECK (user1_id < user2_id),
    
    -- Unique constraint to prevent duplicate friendships
    CONSTRAINT unq_friendship UNIQUE (user1_id, user2_id)
);

-- Indexes for faster friend lookups
CREATE INDEX idx_friends_user1 ON friends(user1_id);
CREATE INDEX idx_friends_user2 ON friends(user2_id);

-- Subscriptions table for one-way update subscriptions
-- subscriber subscribes to updates from target
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    subscriber_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_subscriptions_subscriber FOREIGN KEY (subscriber_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_subscriptions_target FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Prevent self-subscription
    CONSTRAINT chk_no_self_subscription CHECK (subscriber_id != target_id),
    
    -- Unique constraint to prevent duplicate subscriptions
    CONSTRAINT unq_subscription UNIQUE (subscriber_id, target_id)
);

-- Indexes for faster subscription lookups
CREATE INDEX idx_subscriptions_subscriber ON subscriptions(subscriber_id);
CREATE INDEX idx_subscriptions_target ON subscriptions(target_id);

-- Blocks table for blocking updates
-- blocker blocks updates from blocked
CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    blocker_id INTEGER NOT NULL,
    blocked_id INTEGER NOT NULL,
    
    -- Foreign key constraints
    CONSTRAINT fk_blocks_blocker FOREIGN KEY (blocker_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_blocks_blocked FOREIGN KEY (blocked_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Prevent self-blocking
    CONSTRAINT chk_no_self_block CHECK (blocker_id != blocked_id),
    
    -- Unique constraint to prevent duplicate blocks
    CONSTRAINT unq_block UNIQUE (blocker_id, blocked_id)
);

-- Indexes for faster block lookups
CREATE INDEX idx_blocks_blocker ON blocks(blocker_id);
CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);