-- Drop all tables and indexes in reverse order
DROP INDEX IF EXISTS idx_blocks_blocked;
DROP INDEX IF EXISTS idx_blocks_blocker;
DROP TABLE IF EXISTS blocks;

DROP INDEX IF EXISTS idx_subscriptions_target;
DROP INDEX IF EXISTS idx_subscriptions_subscriber;
DROP TABLE IF EXISTS subscriptions;

DROP INDEX IF EXISTS idx_friends_user2;
DROP INDEX IF EXISTS idx_friends_user1;
DROP TABLE IF EXISTS friends;

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;