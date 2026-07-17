-- +goose Up

-- Phase 2.4: Add soft delete support to users, groups, and teams.

ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

ALTER TABLE groups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups (deleted_at);

ALTER TABLE teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_teams_deleted_at ON teams (deleted_at);

-- +goose Down

ALTER TABLE teams DROP INDEX IF EXISTS idx_teams_deleted_at;
ALTER TABLE teams DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE groups DROP INDEX IF EXISTS idx_groups_deleted_at;
ALTER TABLE groups DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE users DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
