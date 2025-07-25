-- +migrate Up
ALTER TABLE users ADD COLUMN account_status TEXT NOT NULL DEFAULT 'free';

-- +migrate Down
ALTER TABLE users DROP COLUMN account_status;