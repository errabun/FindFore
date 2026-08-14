import { useCallback, useEffect, useState } from 'react';
import { groupAdapter } from '../adapters/api/groupAdapter';
import { teeTimeAdapter } from '../adapters/api/teeTimeAdapter';
import { ApiError } from '../adapters/api/httpClient';
import type { Event } from '../domain/teeTime/types';

function messageFrom(err: unknown, fallback: string) {
  return err instanceof ApiError ? err.message : fallback;
}

export function useGroupRounds(groupId: number, hostPlayer: number) {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(Boolean(hostPlayer && groupId));
  const [error, setError] = useState('');

  const refresh = useCallback(() => {
    if (!hostPlayer || !groupId) {
      setEvents([]);
      return Promise.resolve();
    }
    setLoading(true);
    setError('');
    return groupAdapter
      .listEvents(groupId)
      .then(setEvents)
      .catch((err) => setError(messageFrom(err, 'Could not load group rounds')))
      .finally(() => setLoading(false));
  }, [hostPlayer, groupId]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const join = (eventId: number) => {
    return teeTimeAdapter.joinEvent(hostPlayer, eventId).then(() => refresh());
  };

  const cancel = (event: Event) => {
    const op =
      event.host_id === hostPlayer
        ? teeTimeAdapter.deleteEvent(event.id, hostPlayer)
        : teeTimeAdapter.updateInvite(hostPlayer, event.id, 'declined');
    return op.then(() => refresh());
  };

  return { events, loading, error, join, cancel };
}
