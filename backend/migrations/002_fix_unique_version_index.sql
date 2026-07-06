-- +goose Up
DROP INDEX IF EXISTS idx_app_version_name;
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_version_name ON application_versions (application_id, name) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_app_version_name;
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_version_name ON application_versions (application_id, name);
