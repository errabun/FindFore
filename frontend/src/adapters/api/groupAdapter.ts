import type { GroupInvitation, GroupMember, GroupSummary } from '../../domain/group/types';
import type { GroupPort } from '../../ports/groupPort';
import { endpoints, request, requestVoid } from './httpClient';

export const groupAdapter: GroupPort = {
  async listMine() {
    const body = await request<{ groups: GroupSummary[] }>(`${endpoints.groups}?mine=1`);
    return body.groups ?? [];
  },
  async listDiscover(search = '') {
    const q = search ? `&search=${encodeURIComponent(search)}` : '';
    const body = await request<{ groups: GroupSummary[] }>(`${endpoints.groups}?limit=20${q}`);
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
};
