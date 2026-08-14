DROP INDEX IF EXISTS uq_group_invitations_outstanding;
DROP INDEX IF EXISTS idx_group_invitations_group;
DROP INDEX IF EXISTS idx_group_invitations_invitee;
DROP TABLE IF EXISTS group_invitations;

DROP INDEX IF EXISTS idx_group_memberships_group_status;
DROP INDEX IF EXISTS idx_group_memberships_player;
DROP TABLE IF EXISTS group_memberships;

DROP INDEX IF EXISTS idx_groups_privacy;
DROP INDEX IF EXISTS idx_groups_owner;
DROP TABLE IF EXISTS groups;
