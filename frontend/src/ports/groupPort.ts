import type { GroupInvitation, GroupMember, GroupSummary } from '../domain/group/types';
import type { Post } from '../domain/social/types';

export interface GroupPort {
  listMine(): Promise<GroupSummary[]>;
  listDiscover(search?: string, offset?: number): Promise<GroupSummary[]>;
  get(id: number): Promise<GroupSummary>;
  create(name: string, description: string, privacy: 'public' | 'private'): Promise<GroupSummary>;
  update(id: number, name: string, description: string, privacy: 'public' | 'private'): Promise<GroupSummary>;
  delete(id: number): Promise<void>;
  transferOwnership(groupId: number, playerId: number): Promise<GroupSummary>;
  join(id: number): Promise<GroupMember>;
  leave(id: number): Promise<void>;
  listMembers(id: number): Promise<GroupMember[]>;
  removeMember(groupId: number, playerId: number): Promise<void>;
  invite(groupId: number, playerId: number): Promise<unknown>;
  listGroupInvitations(groupId: number): Promise<GroupInvitation[]>;
  cancelInvitation(groupId: number, invitationId: number): Promise<void>;
  listInvitations(): Promise<GroupInvitation[]>;
  acceptInvitation(id: number): Promise<GroupMember>;
  declineInvitation(id: number): Promise<void>;
  listJoinRequests(groupId: number): Promise<GroupMember[]>;
  approveJoinRequest(groupId: number, playerId: number): Promise<GroupMember>;
  denyJoinRequest(groupId: number, playerId: number): Promise<void>;
  listPosts(groupId: number): Promise<Post[]>;
  createPost(groupId: number, body: string): Promise<Post>;
}
