export type GroupPrivacy = 'public' | 'private';
export type GroupRole = 'owner' | 'admin' | 'member';
export type GroupMembershipStatus = 'active' | 'pending';

export interface GroupViewerMembership {
  status: GroupMembershipStatus;
  role: GroupRole;
}

export interface GroupSummary {
  id: number;
  name: string;
  description: string;
  privacy: GroupPrivacy;
  owner: { id: number; name: string };
  member_count: number;
  viewer_membership: GroupViewerMembership | null;
}

export interface GroupMember {
  player_id: number;
  player_name: string;
  role: GroupRole;
  status: GroupMembershipStatus;
}

export interface GroupInvitation {
  id: number;
  group_id: number;
  group_name: string;
  inviter_player_id: number;
  inviter_name: string;
  invitee_player_id: number;
  invitee_name?: string;
  expires_at?: string;
}
