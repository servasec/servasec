-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL    PRIMARY KEY,
    username    VARCHAR(32)  NOT NULL,
    email       VARCHAR(254) NOT NULL,
    password    TEXT         NOT NULL,
    role        TEXT         NOT NULL DEFAULT 'member',
    banned      BOOLEAN      NOT NULL DEFAULT FALSE,
    oauth_provider VARCHAR(50),
    oauth_id    VARCHAR(255),
    avatar_url  VARCHAR(500),
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_oauth_provider_id ON users (oauth_provider, oauth_id);

CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    token_hash  VARCHAR(64)  PRIMARY KEY,
    created_at  TIMESTAMPTZ  NOT NULL
);

CREATE TABLE IF NOT EXISTS groups (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    path        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_path ON groups (path);

CREATE TABLE IF NOT EXISTS scanner_types (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(30)  NOT NULL,
    description VARCHAR(500),
    parser      VARCHAR(100),
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scanner_types_name ON scanner_types (name);

CREATE TABLE IF NOT EXISTS teams (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_name ON teams (name);

CREATE TABLE IF NOT EXISTS applications (
    id               BIGSERIAL    PRIMARY KEY,
    name             VARCHAR(200) NOT NULL,
    description      VARCHAR(1000),
    slug             VARCHAR(100) NOT NULL,
    group_id         BIGINT       NOT NULL,
    repository_url   VARCHAR(500),
    api_token        VARCHAR(64)  NOT NULL,
    asset_criticality VARCHAR(10) DEFAULT 'medium',
    created_at       TIMESTAMPTZ  NOT NULL,
    updated_at       TIMESTAMPTZ  NOT NULL,
    deleted_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_slug ON applications (slug, group_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_applications_api_token ON applications (api_token);
CREATE INDEX IF NOT EXISTS idx_applications_group_id ON applications (group_id);
CREATE INDEX IF NOT EXISTS idx_applications_deleted_at ON applications (deleted_at);

CREATE TABLE IF NOT EXISTS application_versions (
    id              BIGSERIAL    PRIMARY KEY,
    application_id  BIGINT       NOT NULL,
    name            VARCHAR(100) NOT NULL,
    branch          VARCHAR(200),
    tag             VARCHAR(100),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL,
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_app_version_name ON application_versions (application_id, name);
CREATE INDEX IF NOT EXISTS idx_application_versions_deleted_at ON application_versions (deleted_at);

CREATE TABLE IF NOT EXISTS scans (
    id                    BIGSERIAL    PRIMARY KEY,
    application_version_id BIGINT       NOT NULL,
    scanner_type_id       BIGINT       NOT NULL,
    status                VARCHAR(20)  NOT NULL DEFAULT 'pending',
    started_at            TIMESTAMPTZ,
    completed_at          TIMESTAMPTZ,
    raw_results           JSONB,
    created_at            TIMESTAMPTZ  NOT NULL,
    updated_at            TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_scans_application_version_id ON scans (application_version_id);
CREATE INDEX IF NOT EXISTS idx_scans_scanner_type_id ON scans (scanner_type_id);

CREATE TABLE IF NOT EXISTS findings (
    id                    BIGSERIAL    PRIMARY KEY,
    scan_id               BIGINT       NOT NULL,
    application_version_id BIGINT       NOT NULL,
    scanner_type_id       BIGINT       NOT NULL,
    rule_id               VARCHAR(200),
    title                 VARCHAR(500),
    severity              VARCHAR(10)  NOT NULL,
    description           TEXT,
    file_path             VARCHAR(1000),
    line_start            INTEGER,
    line_end              INTEGER,
    cwe_id                VARCHAR(200),
    remediation           TEXT,
    dedupe_hash           VARCHAR(64),
    status                VARCHAR(20)  NOT NULL DEFAULT 'open',
    assigned_to           BIGINT,
    due_date              TIMESTAMPTZ,
    reviewed_by           BIGINT,
    risk_score            DOUBLE PRECISION,
    epss_score            DOUBLE PRECISION,
    fixed_at              TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL,
    updated_at            TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON findings (scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_application_version_id ON findings (application_version_id);
CREATE INDEX IF NOT EXISTS idx_findings_scanner_type_id ON findings (scanner_type_id);
CREATE INDEX IF NOT EXISTS idx_findings_rule_id ON findings (rule_id);
CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings (severity);
CREATE INDEX IF NOT EXISTS idx_findings_dedupe_hash ON findings (dedupe_hash);
CREATE INDEX IF NOT EXISTS idx_findings_status ON findings (status);
CREATE INDEX IF NOT EXISTS idx_findings_assigned_to ON findings (assigned_to);
CREATE INDEX IF NOT EXISTS idx_findings_reviewed_by ON findings (reviewed_by);
CREATE INDEX IF NOT EXISTS idx_findings_risk_score ON findings (risk_score);

CREATE TABLE IF NOT EXISTS team_members (
    id         BIGSERIAL    PRIMARY KEY,
    team_id    BIGINT       NOT NULL,
    user_id    BIGINT       NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_team_user ON team_members (team_id, user_id);

CREATE TABLE IF NOT EXISTS comments (
    id         BIGSERIAL    PRIMARY KEY,
    finding_id BIGINT       NOT NULL,
    user_id    BIGINT       NOT NULL,
    body       TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_comments_finding_id ON comments (finding_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments (user_id);

CREATE TABLE IF NOT EXISTS webhooks (
    id             BIGSERIAL    PRIMARY KEY,
    application_id BIGINT       NOT NULL,
    url            VARCHAR(500) NOT NULL,
    secret         VARCHAR(64),
    events         TEXT,
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ  NOT NULL,
    updated_at     TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_webhooks_application_id ON webhooks (application_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT,
    username   VARCHAR(100),
    action     VARCHAR(10),
    resource   VARCHAR(500),
    status     INTEGER,
    details    TEXT,
    ip         VARCHAR(50),
    created_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs (user_id);

CREATE TABLE IF NOT EXISTS policies (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description VARCHAR(1000),
    scope_type  VARCHAR(20)  NOT NULL,
    scope_value VARCHAR(50)  NOT NULL,
    event_types TEXT         NOT NULL,
    conditions  JSONB,
    actions     JSONB,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    priority    INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL
);

CREATE TABLE IF NOT EXISTS policy_logs (
    id              BIGSERIAL    PRIMARY KEY,
    policy_id       BIGINT       NOT NULL,
    finding_id      BIGINT       NOT NULL,
    event_type      VARCHAR(50),
    conditions_met  BOOLEAN      NOT NULL,
    action_type     VARCHAR(50),
    action_result   VARCHAR(50),
    detail          VARCHAR(500),
    created_at      TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_policy_logs_policy_id ON policy_logs (policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_logs_finding_id ON policy_logs (finding_id);

CREATE TABLE IF NOT EXISTS user_api_keys (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL,
    name         VARCHAR(100) NOT NULL,
    key_hash     VARCHAR(64)  NOT NULL,
    key_prefix   VARCHAR(20)  NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_api_keys_key_hash ON user_api_keys (key_hash);
CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_id ON user_api_keys (user_id);

CREATE TABLE IF NOT EXISTS casbin_rules (
    id    BIGSERIAL    PRIMARY KEY,
    ptype VARCHAR(100),
    v0    VARCHAR(100),
    v1    VARCHAR(100),
    v2    VARCHAR(100),
    v3    VARCHAR(100),
    v4    VARCHAR(100),
    v5    VARCHAR(100)
);

-- +goose Down

DROP TABLE IF EXISTS casbin_rules;
DROP TABLE IF EXISTS user_api_keys;
DROP TABLE IF EXISTS policy_logs;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS findings;
DROP TABLE IF EXISTS scans;
DROP TABLE IF EXISTS application_versions;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS scanner_types;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS blacklisted_tokens;
DROP TABLE IF EXISTS users;
