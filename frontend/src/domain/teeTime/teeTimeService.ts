import type { Event } from './types';

export function filterCommitted(events: Event[], playerId: number): Event[] {
  return events.filter((event) => event.accepted.includes(playerId));
}

export function filterAvailable(events: Event[], playerId: number): Event[] {
  return events.filter((event) => {
    if (
      event.declined.includes(playerId) ||
      event.accepted.includes(playerId) ||
      event.closed.includes(playerId)
    ) {
      return false;
    }
    return event.pending.includes(playerId) || !event.private;
  });
}

export function filterPublicInvites(events: Event[]): Event[] {
  return events.filter((event) => !event.private);
}

/** Invites hosted by someone in the viewer's social graph (either direction). */
export function filterFriendInvites(
  events: Event[],
  connectedHostIds: number[],
): Event[] {
  return events.filter((event) => connectedHostIds.includes(event.host_id));
}

/**
 * Friendships are stored one-way (follower → followee). For dashboard filtering,
 * treat either direction as a connection so "Eric added Ashley" still surfaces
 * Eric's tee times on Ashley's Friends segment.
 */
export function buildConnectedPlayerIds(
  currentUserId: number,
  followingIds: number[],
  allPlayers: { id: number; friends: number[] }[],
): number[] {
  const followers = allPlayers
    .filter(
      (p) => p.id !== currentUserId && p.friends.includes(currentUserId),
    )
    .map((p) => p.id);
  return [...new Set([...followingIds, ...followers])];
}

export function isHost(event: Event, playerId: number): boolean {
  return event.host_id === playerId;
}
