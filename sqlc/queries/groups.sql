-- name: GetGroupByID :one
SELECT id, owner_player_id, name, description, privacy, created_at, updated_at
FROM groups
WHERE id = $1;

-- name: InsertGroup :one
INSERT INTO groups (owner_player_id, name, description, privacy, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, owner_player_id, name, description, privacy, created_at, updated_at;

-- name: UpdateGroup :one
UPDATE groups
SET name = $2,
    description = $3,
    privacy = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING id, owner_player_id, name, description, privacy, created_at, updated_at;

-- name: UpdateGroupOwner :exec
UPDATE groups
SET owner_player_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE id = $1;

-- name: ListPublicGroups :many
SELECT id, owner_player_id, name, description, privacy, created_at, updated_at
FROM groups
WHERE privacy = 'public'
  AND (sqlc.arg(search)::text = '' OR name ILIKE '%' || sqlc.arg(search)::text || '%')
ORDER BY name ASC, id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListGroupsByPlayer :many
SELECT g.id, g.owner_player_id, g.name, g.description, g.privacy, g.created_at, g.updated_at
FROM groups g
INNER JOIN group_memberships m ON m.group_id = g.id
WHERE m.player_id = $1
  AND m.status = 'active'
ORDER BY g.name ASC, g.id ASC
LIMIT $2 OFFSET $3;

-- name: ListPublicGroupSummaries :many
SELECT
    g.id,
    g.owner_player_id,
    g.name,
    g.description,
    g.privacy,
    g.created_at,
    g.updated_at,
    COALESCE(owner.name, '') AS owner_name,
    (
        SELECT COUNT(*)::bigint
        FROM group_memberships cm
        WHERE cm.group_id = g.id
          AND cm.status = 'active'
    ) AS member_count,
    viewer.role AS viewer_role,
    viewer.status AS viewer_status,
    viewer.created_at AS viewer_created_at,
    viewer.updated_at AS viewer_updated_at
FROM groups g
INNER JOIN players owner ON owner.id = g.owner_player_id
LEFT JOIN group_memberships viewer
    ON viewer.group_id = g.id
   AND viewer.player_id = sqlc.arg(player_id)
WHERE g.privacy = 'public'
  AND (sqlc.arg(search)::text = '' OR g.name ILIKE '%' || sqlc.arg(search)::text || '%')
ORDER BY g.name ASC, g.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListGroupSummariesByPlayer :many
SELECT
    g.id,
    g.owner_player_id,
    g.name,
    g.description,
    g.privacy,
    g.created_at,
    g.updated_at,
    COALESCE(owner.name, '') AS owner_name,
    (
        SELECT COUNT(*)::bigint
        FROM group_memberships cm
        WHERE cm.group_id = g.id
          AND cm.status = 'active'
    ) AS member_count,
    m.role AS viewer_role,
    m.status AS viewer_status,
    m.created_at AS viewer_created_at,
    m.updated_at AS viewer_updated_at
FROM groups g
INNER JOIN group_memberships m ON m.group_id = g.id
INNER JOIN players owner ON owner.id = g.owner_player_id
WHERE m.player_id = sqlc.arg(player_id)
  AND m.status = 'active'
ORDER BY g.name ASC, g.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountActiveGroupMembers :one
SELECT COUNT(*)::bigint
FROM group_memberships
WHERE group_id = $1
  AND status = 'active';

-- name: GetGroupMembership :one
SELECT group_id, player_id, role, status, created_at, updated_at
FROM group_memberships
WHERE group_id = $1
  AND player_id = $2;

-- name: ListActiveGroupMembers :many
SELECT m.group_id, m.player_id, m.role, m.status, m.created_at, m.updated_at, p.name AS player_name
FROM group_memberships m
INNER JOIN players p ON p.id = m.player_id
WHERE m.group_id = $1
  AND m.status = 'active'
ORDER BY
    CASE m.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END,
    p.name ASC,
    m.player_id ASC
LIMIT $2 OFFSET $3;

-- name: ListPendingGroupMembers :many
SELECT m.group_id, m.player_id, m.role, m.status, m.created_at, m.updated_at, p.name AS player_name
FROM group_memberships m
INNER JOIN players p ON p.id = m.player_id
WHERE m.group_id = $1
  AND m.status = 'pending'
ORDER BY m.created_at ASC, m.player_id ASC;

-- name: InsertGroupMembership :one
INSERT INTO group_memberships (group_id, player_id, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING group_id, player_id, role, status, created_at, updated_at;

-- name: UpdateGroupMembership :one
UPDATE group_memberships
SET role = $3,
    status = $4,
    updated_at = NOW()
WHERE group_id = $1
  AND player_id = $2
RETURNING group_id, player_id, role, status, created_at, updated_at;

-- name: DeleteGroupMembership :exec
DELETE FROM group_memberships
WHERE group_id = $1
  AND player_id = $2;

-- name: GetGroupInvitationByID :one
SELECT id, group_id, inviter_player_id, invitee_player_id, created_at, expires_at, accepted_at, declined_at
FROM group_invitations
WHERE id = $1;

-- name: GetOutstandingGroupInvitation :one
SELECT id, group_id, inviter_player_id, invitee_player_id, created_at, expires_at, accepted_at, declined_at
FROM group_invitations
WHERE group_id = $1
  AND invitee_player_id = $2
  AND accepted_at IS NULL
  AND declined_at IS NULL;

-- name: ListGroupInvitationsByInvitee :many
SELECT i.id, i.group_id, i.inviter_player_id, i.invitee_player_id,
       i.created_at, i.expires_at, i.accepted_at, i.declined_at,
       g.name AS group_name, p.name AS inviter_name
FROM group_invitations i
INNER JOIN groups g ON g.id = i.group_id
INNER JOIN players p ON p.id = i.inviter_player_id
WHERE i.invitee_player_id = $1
  AND i.accepted_at IS NULL
  AND i.declined_at IS NULL
ORDER BY i.created_at DESC;

-- name: ListOutstandingGroupInvitations :many
SELECT i.id, i.group_id, i.inviter_player_id, i.invitee_player_id,
       i.created_at, i.expires_at, i.accepted_at, i.declined_at,
       g.name AS group_name, inviter.name AS inviter_name, invitee.name AS invitee_name
FROM group_invitations i
INNER JOIN groups g ON g.id = i.group_id
INNER JOIN players inviter ON inviter.id = i.inviter_player_id
INNER JOIN players invitee ON invitee.id = i.invitee_player_id
WHERE i.group_id = $1
  AND i.accepted_at IS NULL
  AND i.declined_at IS NULL
ORDER BY i.created_at DESC, i.id DESC;

-- name: InsertGroupInvitation :one
INSERT INTO group_invitations (
    group_id, inviter_player_id, invitee_player_id, created_at, expires_at
)
VALUES ($1, $2, $3, NOW(), $4)
RETURNING id, group_id, inviter_player_id, invitee_player_id, created_at, expires_at, accepted_at, declined_at;

-- name: MarkGroupInvitationAccepted :one
UPDATE group_invitations
SET accepted_at = NOW()
WHERE id = $1
  AND accepted_at IS NULL
  AND declined_at IS NULL
RETURNING id, group_id, inviter_player_id, invitee_player_id, created_at, expires_at, accepted_at, declined_at;

-- name: MarkGroupInvitationDeclined :one
UPDATE group_invitations
SET declined_at = NOW()
WHERE id = $1
  AND accepted_at IS NULL
  AND declined_at IS NULL
RETURNING id, group_id, inviter_player_id, invitee_player_id, created_at, expires_at, accepted_at, declined_at;
