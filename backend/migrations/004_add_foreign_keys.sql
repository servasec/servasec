-- +goose Up

-- Phase 1.2: Add foreign key constraints to enforce referential integrity at the database level.

ALTER TABLE applications
    ADD CONSTRAINT fk_applications_group_id
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE application_versions
    ADD CONSTRAINT fk_application_versions_application_id
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE scans
    ADD CONSTRAINT fk_scans_application_version_id
    FOREIGN KEY (application_version_id) REFERENCES application_versions(id) ON DELETE CASCADE;

ALTER TABLE scans
    ADD CONSTRAINT fk_scans_scanner_type_id
    FOREIGN KEY (scanner_type_id) REFERENCES scanner_types(id) ON DELETE RESTRICT;

ALTER TABLE findings
    ADD CONSTRAINT fk_findings_scan_id
    FOREIGN KEY (scan_id) REFERENCES scans(id) ON DELETE CASCADE;

ALTER TABLE findings
    ADD CONSTRAINT fk_findings_application_version_id
    FOREIGN KEY (application_version_id) REFERENCES application_versions(id) ON DELETE CASCADE;

ALTER TABLE findings
    ADD CONSTRAINT fk_findings_scanner_type_id
    FOREIGN KEY (scanner_type_id) REFERENCES scanner_types(id) ON DELETE RESTRICT;

ALTER TABLE findings
    ADD CONSTRAINT fk_findings_assigned_to
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE findings
    ADD CONSTRAINT fk_findings_reviewed_by
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE team_members
    ADD CONSTRAINT fk_team_members_team_id
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

ALTER TABLE team_members
    ADD CONSTRAINT fk_team_members_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE comments
    ADD CONSTRAINT fk_comments_finding_id
    FOREIGN KEY (finding_id) REFERENCES findings(id) ON DELETE CASCADE;

ALTER TABLE comments
    ADD CONSTRAINT fk_comments_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE webhooks
    ADD CONSTRAINT fk_webhooks_application_id
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE policy_logs
    ADD CONSTRAINT fk_policy_logs_policy_id
    FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE;

ALTER TABLE policy_logs
    ADD CONSTRAINT fk_policy_logs_finding_id
    FOREIGN KEY (finding_id) REFERENCES findings(id) ON DELETE SET NULL;

ALTER TABLE user_api_keys
    ADD CONSTRAINT fk_user_api_keys_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down

ALTER TABLE user_api_keys DROP CONSTRAINT IF EXISTS fk_user_api_keys_user_id;
ALTER TABLE policy_logs DROP CONSTRAINT IF EXISTS fk_policy_logs_finding_id;
ALTER TABLE policy_logs DROP CONSTRAINT IF EXISTS fk_policy_logs_policy_id;
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS fk_webhooks_application_id;
ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_user_id;
ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_finding_id;
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS fk_team_members_user_id;
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS fk_team_members_team_id;
ALTER TABLE findings DROP CONSTRAINT IF EXISTS fk_findings_reviewed_by;
ALTER TABLE findings DROP CONSTRAINT IF EXISTS fk_findings_assigned_to;
ALTER TABLE findings DROP CONSTRAINT IF EXISTS fk_findings_scanner_type_id;
ALTER TABLE findings DROP CONSTRAINT IF EXISTS fk_findings_application_version_id;
ALTER TABLE findings DROP CONSTRAINT IF EXISTS fk_findings_scan_id;
ALTER TABLE scans DROP CONSTRAINT IF EXISTS fk_scans_scanner_type_id;
ALTER TABLE scans DROP CONSTRAINT IF EXISTS fk_scans_application_version_id;
ALTER TABLE application_versions DROP CONSTRAINT IF EXISTS fk_application_versions_application_id;
ALTER TABLE applications DROP CONSTRAINT IF EXISTS fk_applications_group_id;
