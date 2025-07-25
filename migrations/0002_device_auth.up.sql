-- +migrate Up
ALTER TABLE users ADD COLUMN device_id TEXT;
ALTER TABLE users ADD COLUMN reset_requested BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN reset_token TEXT;
ALTER TABLE users ADD COLUMN reset_token_expiry TIMESTAMP;

-- +migrate Down
ALTER TABLE users DROP COLUMN device_id;
ALTER TABLE users DROP COLUMN reset_requested;
ALTER TABLE users DROP COLUMN reset_token;
ALTER TABLE users DROP COLUMN reset_token_expiry;