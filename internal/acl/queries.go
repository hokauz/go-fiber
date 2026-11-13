package acl

const (
	// Groups

	queryInsertGroup = `
INSERT INTO acl_groups (id, resource_type, resource_id, name, access_level, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now(), now())
RETURNING id, resource_type, resource_id, name, access_level, created_by, created_at, updated_at, deleted_at`

	queryGetGroupByID = `
SELECT id, resource_type, resource_id, name, access_level, created_by, created_at, updated_at, deleted_at
FROM acl_groups
WHERE id = $1 AND deleted_at IS NULL`

	queryListGroupsByResource = `
SELECT id, resource_type, resource_id, name, access_level, created_by, created_at, updated_at, deleted_at
FROM acl_groups
WHERE resource_type = $1 AND resource_id = $2 AND deleted_at IS NULL
ORDER BY created_at ASC`

	queryUpdateGroup = `
UPDATE acl_groups SET
	name = $2,
	access_level = $3,
	updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, resource_type, resource_id, name, access_level, created_by, created_at, updated_at, deleted_at`

	querySoftDeleteGroup = `
UPDATE acl_groups SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL`

	queryCancelPendingInvitesByGroup = `
UPDATE acl_invitations
SET status = 'cancelled', updated_at = now()
WHERE group_id = $1 AND status = 'pending'`

	querySoftRemoveMembersByGroup = `
UPDATE acl_group_members SET removed_at = now()
WHERE group_id = $1 AND removed_at IS NULL`

	// Members

	queryInsertMember = `
INSERT INTO acl_group_members (id, group_id, user_id, source_type, source_id, created_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING id, group_id, user_id, source_type, source_id, created_at, removed_at`

	queryInsertMemberTx = `
INSERT INTO acl_group_members (id, group_id, user_id, source_type, source_id, created_at)
VALUES ($1, $2, $3, 'invite', $4, now())`

	queryListMembersByGroup = `
SELECT id, group_id, user_id, source_type, source_id, created_at, removed_at
FROM acl_group_members
WHERE group_id = $1 AND removed_at IS NULL
ORDER BY created_at ASC`

	queryGetMemberByID = `
SELECT id, group_id, user_id, source_type, source_id, created_at, removed_at
FROM acl_group_members
WHERE id = $1 AND removed_at IS NULL`

	querySoftRemoveMember = `
UPDATE acl_group_members SET removed_at = now()
WHERE id = $1 AND group_id = $2 AND removed_at IS NULL`

	// Invitations

	queryInsertInvitation = `
INSERT INTO acl_invitations (id, group_id, created_by, code_hash, code_expires_at, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, 'pending', now(), now())
RETURNING id, group_id, created_by, code_hash, code_expires_at, status, consumed_at, accepted_by, created_at, updated_at`

	queryGetInvitationByCodeHash = `
SELECT id, group_id, created_by, code_hash, code_expires_at, status, consumed_at, accepted_by, created_at, updated_at
FROM acl_invitations
WHERE code_hash = $1`

	queryGetInvitationByCodeHashForUpdate = `
SELECT id, group_id, created_by, code_hash, code_expires_at, status, consumed_at, accepted_by, created_at, updated_at
FROM acl_invitations
WHERE code_hash = $1
FOR UPDATE`

	queryGetInvitationByID = `
SELECT id, group_id, created_by, code_hash, code_expires_at, status, consumed_at, accepted_by, created_at, updated_at
FROM acl_invitations
WHERE id = $1`

	queryGetInvitationByIDForUpdate = `
SELECT id, group_id, created_by, code_hash, code_expires_at, status, consumed_at, accepted_by, created_at, updated_at
FROM acl_invitations
WHERE id = $1
FOR UPDATE`

	queryListInvitationsByResource = `
SELECT i.id, i.group_id, i.created_by, i.code_hash, i.code_expires_at, i.status, i.consumed_at, i.accepted_by, i.created_at, i.updated_at
FROM acl_invitations i
JOIN acl_groups g ON g.id = i.group_id
WHERE g.resource_type = $1 AND g.resource_id = $2
ORDER BY i.created_at DESC`

	queryAcceptInvitation = `
UPDATE acl_invitations
SET status = 'accepted', consumed_at = now(), accepted_by = $2, updated_at = now()
WHERE id = $1`

	queryRejectInvitation = `
UPDATE acl_invitations
SET status = 'rejected', updated_at = now()
WHERE id = $1`

	queryCancelInvitation = `
UPDATE acl_invitations
SET status = 'cancelled', updated_at = now()
WHERE id = $1 AND status = 'pending'`

	// Authorization

	queryGetEffectiveAccessLevel = `
SELECT g.access_level
FROM acl_group_members m
JOIN acl_groups g ON g.id = m.group_id
WHERE g.resource_type = $1
  AND g.resource_id = $2
  AND m.user_id = $3
  AND m.removed_at IS NULL
  AND g.deleted_at IS NULL
ORDER BY
  CASE g.access_level
    WHEN 'manage_access' THEN 1
    WHEN 'edit'          THEN 2
    WHEN 'view'          THEN 3
  END
LIMIT 1`

	queryListAccessibleResourceAccessLevels = `
SELECT g.resource_id,
  CASE MIN(CASE g.access_level
    WHEN 'manage_access' THEN 1
    WHEN 'edit'          THEN 2
    WHEN 'view'          THEN 3
  END)
    WHEN 1 THEN 'manage_access'
    WHEN 2 THEN 'edit'
    WHEN 3 THEN 'view'
  END AS effective_access_level
FROM acl_group_members m
JOIN acl_groups g ON g.id = m.group_id
WHERE g.resource_type = $1
  AND m.user_id = $2
  AND m.removed_at IS NULL
  AND g.deleted_at IS NULL
GROUP BY g.resource_id`

	// Delete requests

	queryInsertDeleteRequest = `
INSERT INTO acl_delete_requests (
	id, resource_type, resource_id, requested_by, status, reason, expires_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
RETURNING id, resource_type, resource_id, requested_by, status, reason, created_at, updated_at, approved_by, approved_at, rejected_by, rejected_at, executed_at, expires_at`

	queryGetDeleteRequestByID = `
SELECT id, resource_type, resource_id, requested_by, status, reason, created_at, updated_at, approved_by, approved_at, rejected_by, rejected_at, executed_at, expires_at
FROM acl_delete_requests
WHERE id = $1`

	queryGetDeleteRequestByIDForUpdate = `
SELECT id, resource_type, resource_id, requested_by, status, reason, created_at, updated_at, approved_by, approved_at, rejected_by, rejected_at, executed_at, expires_at
FROM acl_delete_requests
WHERE id = $1
FOR UPDATE`

	queryListDeleteRequestsByResource = `
SELECT id, resource_type, resource_id, requested_by, status, reason, created_at, updated_at, approved_by, approved_at, rejected_by, rejected_at, executed_at, expires_at
FROM acl_delete_requests
WHERE resource_type = $1
  AND resource_id = $2
  AND ($3 IS NULL OR requested_by = $3)
  AND ($4 IS NULL OR status = $4)
ORDER BY created_at DESC
LIMIT $5`

	queryApproveDeleteRequest = `
UPDATE acl_delete_requests
SET status = 'approved', approved_by = $2, approved_at = now(), updated_at = now()
WHERE id = $1 AND status = 'pending'`

	queryRejectDeleteRequest = `
UPDATE acl_delete_requests
SET status = 'rejected', rejected_by = $2, rejected_at = now(), updated_at = now()
WHERE id = $1 AND status = 'pending'`

	queryMarkDeleteRequestExecuted = `
UPDATE acl_delete_requests
SET status = 'executed', executed_at = now(), updated_at = now()
WHERE id = $1 AND status IN ('approved', 'pending')`

	// Activity logs

	queryInsertActivityLog = `
INSERT INTO acl_activity_logs (
	id, actor_user_id, entity_type, entity_id, action, resource_type, resource_id, metadata, correlation_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())`

	queryListActivityLogsByResource = `
SELECT id, actor_user_id, entity_type, entity_id, action, resource_type, resource_id, metadata, correlation_id, created_at
FROM acl_activity_logs
WHERE resource_type = $1
  AND resource_id = $2
  AND ($3 IS NULL OR actor_user_id = $3)
  AND ($4 IS NULL OR action = $4)
  AND ($5::timestamptz IS NULL OR created_at >= $5)
  AND ($6::timestamptz IS NULL OR created_at <= $6)
ORDER BY created_at DESC
LIMIT $7`
)
