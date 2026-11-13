-- +migrate Up

CREATE TABLE acl_groups (
    id              UUID        PRIMARY KEY,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     UUID        NOT NULL,
    name            VARCHAR(80) NOT NULL,
    access_level    VARCHAR(20) NOT NULL,
    created_by      UUID        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT acl_groups_access_level_check
        CHECK (access_level IN ('manage_access', 'edit', 'view')),
    CONSTRAINT uq_acl_groups_resource_name
        UNIQUE NULLS NOT DISTINCT (resource_type, resource_id, name, deleted_at)
);

CREATE INDEX idx_acl_groups_resource
    ON acl_groups (resource_type, resource_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_acl_groups_created_by
    ON acl_groups (created_by)
    WHERE deleted_at IS NULL;

CREATE TABLE acl_group_members (
    id          UUID        PRIMARY KEY,
    group_id    UUID        NOT NULL REFERENCES acl_groups(id),
    user_id     UUID        NOT NULL,
    source_type VARCHAR(20) NOT NULL DEFAULT 'invite',
    source_id   UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    removed_at  TIMESTAMPTZ,

    CONSTRAINT acl_group_members_source_check
        CHECK (source_type IN ('invite', 'direct'))
);

CREATE UNIQUE INDEX uq_acl_group_members_active
    ON acl_group_members (group_id, user_id)
    WHERE removed_at IS NULL;

CREATE INDEX idx_acl_group_members_user
    ON acl_group_members (user_id)
    WHERE removed_at IS NULL;

CREATE TABLE acl_invitations (
    id               UUID        PRIMARY KEY,
    group_id         UUID        NOT NULL REFERENCES acl_groups(id),
    created_by       UUID        NOT NULL,
    code_hash        VARCHAR(64) NOT NULL UNIQUE,
    code_expires_at  TIMESTAMPTZ NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    consumed_at      TIMESTAMPTZ,
    accepted_by      UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT acl_invitations_status_check
        CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled', 'expired'))
);

CREATE INDEX idx_acl_invitations_group
    ON acl_invitations (group_id)
    WHERE status = 'pending';

CREATE INDEX idx_acl_invitations_created_by
    ON acl_invitations (created_by);

CREATE TABLE acl_resource_owners (
    resource_type  VARCHAR(50) NOT NULL,
    resource_id    UUID        NOT NULL,
    owner_user_id  UUID        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (resource_type, resource_id)
);

CREATE INDEX idx_acl_resource_owners_owner
    ON acl_resource_owners (owner_user_id);

CREATE TABLE acl_activity_logs (
    id             UUID        PRIMARY KEY,
    actor_user_id  UUID,
    entity_type    VARCHAR(30) NOT NULL,
    entity_id      UUID        NOT NULL,
    action         VARCHAR(30) NOT NULL,
    resource_type  VARCHAR(50) NOT NULL,
    resource_id    UUID        NOT NULL,
    metadata       JSONB       NOT NULL DEFAULT '{}',
    correlation_id VARCHAR(64),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_acl_activity_logs_resource
    ON acl_activity_logs (resource_type, resource_id);

CREATE INDEX idx_acl_activity_logs_actor
    ON acl_activity_logs (actor_user_id)
    WHERE actor_user_id IS NOT NULL;

CREATE INDEX idx_acl_activity_logs_correlation_id
    ON acl_activity_logs (correlation_id)
    WHERE correlation_id IS NOT NULL;

CREATE TABLE acl_delete_requests (
    id             UUID        PRIMARY KEY,
    resource_type  VARCHAR(50) NOT NULL,
    resource_id    UUID        NOT NULL,
    requested_by   UUID        NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    reason         TEXT,
    approved_by    UUID,
    approved_at    TIMESTAMPTZ,
    rejected_by    UUID,
    rejected_at    TIMESTAMPTZ,
    executed_at    TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT acl_delete_requests_status_check
        CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'executed'))
);

CREATE INDEX idx_acl_delete_requests_resource
    ON acl_delete_requests (resource_type, resource_id);

CREATE INDEX idx_acl_delete_requests_requested_by
    ON acl_delete_requests (requested_by);

CREATE UNIQUE INDEX uq_acl_delete_requests_open_resource
    ON acl_delete_requests (resource_type, resource_id)
    WHERE status IN ('pending', 'approved');

-- +migrate Down

DROP INDEX IF EXISTS uq_acl_delete_requests_open_resource;
DROP TABLE IF EXISTS acl_delete_requests;
DROP INDEX IF EXISTS idx_acl_activity_logs_correlation_id;
DROP TABLE IF EXISTS acl_activity_logs;
DROP TABLE IF EXISTS acl_resource_owners;
DROP TABLE IF EXISTS acl_invitations;
DROP INDEX IF EXISTS uq_acl_group_members_active;
DROP TABLE IF EXISTS acl_group_members;
DROP TABLE IF EXISTS acl_groups;
