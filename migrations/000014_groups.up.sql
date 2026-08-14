-- Persistent golfer groups: community, membership, and invitations.
-- v1 roles are owner/member; admin is reserved in the CHECK for a later slice.

CREATE TABLE groups (
    id BIGSERIAL PRIMARY KEY,
    owner_player_id BIGINT NOT NULL REFERENCES players(id),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(1000) NOT NULL DEFAULT '',
    privacy VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_groups_privacy CHECK (privacy IN ('public', 'private')),
    CONSTRAINT chk_groups_name_not_blank CHECK (length(trim(name)) > 0)
);

CREATE INDEX idx_groups_owner ON groups (owner_player_id);
CREATE INDEX idx_groups_privacy ON groups (privacy);

CREATE TABLE group_memberships (
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, player_id),
    CONSTRAINT chk_group_memberships_role CHECK (role IN ('owner', 'admin', 'member')),
    CONSTRAINT chk_group_memberships_status CHECK (status IN ('active', 'pending'))
);

CREATE INDEX idx_group_memberships_player ON group_memberships (player_id);
CREATE INDEX idx_group_memberships_group_status ON group_memberships (group_id, status);

CREATE TABLE group_invitations (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    inviter_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    invitee_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    accepted_at TIMESTAMPTZ,
    declined_at TIMESTAMPTZ,
    CONSTRAINT chk_group_invitations_not_self CHECK (inviter_player_id <> invitee_player_id)
);

CREATE INDEX idx_group_invitations_invitee ON group_invitations (invitee_player_id);
CREATE INDEX idx_group_invitations_group ON group_invitations (group_id);

CREATE UNIQUE INDEX uq_group_invitations_outstanding
    ON group_invitations (group_id, invitee_player_id)
    WHERE accepted_at IS NULL AND declined_at IS NULL;
