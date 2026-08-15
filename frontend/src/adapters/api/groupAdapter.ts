import type { GroupChatSession, GroupInvitation, GroupMember, GroupSummary } from '../../domain/group/types';
import type { Post } from '../../domain/social/types';
import type { Event } from '../../domain/teeTime/types';
import type { GroupPort } from '../../ports/groupPort';
import { endpoints, request, requestVoid } from './httpClient';

const DISCOVER_LIMIT = 20;

export const groupAdapter: GroupPort = {
  async listMine() {
    const body = await request<{ groups: GroupSummary[] }>(`${endpoints.groups}?mine=1`);
    return body.groups ?? [];
  },
  async listDiscover(search = '', offset = 0) {
    const params = new URLSearchParams({ limit: String(DISCOVER_LIMIT) });
    if (search) params.set('search', search);
    if (offset) params.set('offset', String(offset));
    const body = await request<{ groups: GroupSummary[] }>(`${endpoints.groups}?${params.toString()}`);
    return body.groups ?? [];
  },
  get(id) {
    return request<GroupSummary>(`${endpoints.groups}/${id}`);
  },
  create(name, description, privacy) {
    return request<GroupSummary>(endpoints.groups, {
      method: 'POST',
      body: JSON.stringify({ name, description, privacy }),
    });
  },
  update(id, name, description, privacy) {
    return request<GroupSummary>(`${endpoints.groups}/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ name, description, privacy }),
    });
  },
  delete(id) {
    return requestVoid(`${endpoints.groups}/${id}`, { method: 'DELETE' });
  },
  transferOwnership(groupId, playerId) {
    return request<GroupSummary>(`${endpoints.groups}/${groupId}/transfer-ownership`, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId }),
    });
  },
  join(id) {
    return request<GroupMember>(`${endpoints.groups}/${id}/join`, { method: 'POST' });
  },
  leave(id) {
    return requestVoid(`${endpoints.groups}/${id}/leave`, { method: 'POST' });
  },
  async listMembers(id) {
    const body = await request<{ members: GroupMember[] }>(`${endpoints.groups}/${id}/members`);
    return body.members ?? [];
  },
  removeMember(groupId, playerId) {
    return requestVoid(`${endpoints.groups}/${groupId}/members/${playerId}`, { method: 'DELETE' });
  },
  invite(groupId, playerId) {
    return request(`${endpoints.groups}/${groupId}/invitations`, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId }),
    });
  },
  async listGroupInvitations(groupId) {
    const body = await request<{ invitations: GroupInvitation[] }>(
      `${endpoints.groups}/${groupId}/invitations`,
    );
    return body.invitations ?? [];
  },
  cancelInvitation(groupId, invitationId) {
    return requestVoid(`${endpoints.groups}/${groupId}/invitations/${invitationId}`, {
      method: 'DELETE',
    });
  },
  async listInvitations() {
    const body = await request<{ invitations: GroupInvitation[] }>(endpoints.groupInvitations);
    return body.invitations ?? [];
  },
  acceptInvitation(id) {
    return request<GroupMember>(`${endpoints.groupInvitations}/${id}/accept`, { method: 'POST' });
  },
  declineInvitation(id) {
    return requestVoid(`${endpoints.groupInvitations}/${id}/decline`, { method: 'POST' });
  },
  async listJoinRequests(groupId) {
    const body = await request<{ join_requests: GroupMember[] }>(`${endpoints.groups}/${groupId}/join-requests`);
    return body.join_requests ?? [];
  },
  approveJoinRequest(groupId, playerId) {
    return request<GroupMember>(`${endpoints.groups}/${groupId}/join-requests/${playerId}/approve`, { method: 'POST' });
  },
  denyJoinRequest(groupId, playerId) {
    return requestVoid(`${endpoints.groups}/${groupId}/join-requests/${playerId}/deny`, { method: 'POST' });
  },
  async listPosts(groupId) {
    const body = await request<{ posts: Post[] }>(`${endpoints.groups}/${groupId}/posts`);
    return body.posts ?? [];
  },
  createPost(groupId, body) {
    return request<Post>(`${endpoints.groups}/${groupId}/posts`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    });
  },
  async listEvents(groupId) {
    const body = await request<{ events: Event[] }>(`${endpoints.groups}/${groupId}/events`);
    return body.events ?? [];
  },
  getChat(groupId) {
    return request<GroupChatSession>(`${endpoints.groups}/${groupId}/chat`);
  },
};
