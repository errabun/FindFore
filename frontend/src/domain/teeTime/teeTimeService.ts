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

export function isHost(event: Event, playerId: number): boolean {
  return event.host_id === playerId;
}
