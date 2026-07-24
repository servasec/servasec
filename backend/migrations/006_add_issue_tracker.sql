-- +goose Up
CREATE TABLE IF NOT EXISTS issue_trackers (
    id BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    auth_type VARCHAR(20) NOT NULL DEFAULT 'pat',
    encrypted_token VARCHAR(500),
    encrypted_token_iv VARCHAR(44),
    github_app_id BIGINT,
    github_installation_id BIGINT,
    encrypted_github_app_key VARCHAR(6000),
    encrypted_github_app_key_iv VARCHAR(44),
    token_expires_at TIMESTAMPTZ,
    severity_threshold VARCHAR(10) NOT NULL DEFAULT 'high',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_issue_trackers_deleted_at ON issue_trackers(deleted_at);

CREATE TABLE IF NOT EXISTS issue_tracker_issues (
    id BIGSERIAL PRIMARY KEY,
    issue_tracker_id BIGINT NOT NULL REFERENCES issue_trackers(id) ON DELETE CASCADE,
    finding_id BIGINT NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    external_issue_id VARCHAR(100) NOT NULL,
    external_issue_url VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT idx_tracker_finding UNIQUE (issue_tracker_id, finding_id)
);

-- +goose Down
DROP TABLE IF EXISTS issue_tracker_issues;
DROP TABLE IF EXISTS issue_trackers;
