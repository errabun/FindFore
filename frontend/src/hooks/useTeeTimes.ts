import { useState, useEffect, useCallback } from 'react';
import type { Event } from '../domain/teeTime/types';
import { teeTimeAdapter } from '../adapters/api/teeTimeAdapter';

export function useTeeTimes(hostPlayer: number) {
  const [events, setEvents] = useState<Event[]>([]);
  const [friendsEvents, setFriendsEvents] = useState<Event[]>([]);
  const [groupJoinableEvents, setGroupJoinableEvents] = useState<Event[]>([]);

  const refreshGroupJoinable = useCallback(() => {
    if (!hostPlayer) {
      setGroupJoinableEvents([]);
      return;
    }
    teeTimeAdapter.getGroupJoinableEvents().then(setGroupJoinableEvents);
  }, [hostPlayer]);

  const refreshEvents = useCallback(() => {
    if (!hostPlayer) return;
    teeTimeAdapter.getEvents(hostPlayer).then(setEvents);
    teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
    refreshGroupJoinable();
  }, [hostPlayer, refreshGroupJoinable]);

  const updateInvite = (eventId: number, status: string) => {
    teeTimeAdapter.updateInvite(hostPlayer, eventId, status).then((events) => {
      setEvents(events);
      teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
      refreshGroupJoinable();
    });
  };

  const cancelCommitment = (event: Event) => {
    if (event.host_id === hostPlayer) {
      teeTimeAdapter.deleteEvent(event.id, hostPlayer).then((events) => {
        setEvents(events);
        teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
        refreshGroupJoinable();
      });
    } else {
      teeTimeAdapter.updateInvite(hostPlayer, event.id, 'declined').then((events) => {
        setEvents(events);
        teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
        refreshGroupJoinable();
      });
    }
  };

  const joinTeeTime = (eventId: number) => {
    teeTimeAdapter.joinEvent(hostPlayer, eventId).then((events) => {
      setEvents(events);
      setFriendsEvents((prev) => prev.filter((e) => e.id !== eventId));
      setGroupJoinableEvents((prev) => prev.filter((e) => e.id !== eventId));
    });
  };

  useEffect(() => {
    if (hostPlayer) {
      teeTimeAdapter.getEvents(hostPlayer).then(setEvents);
      teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
      teeTimeAdapter.getGroupJoinableEvents().then(setGroupJoinableEvents);
    } else {
      setEvents([]);
      setFriendsEvents([]);
      setGroupJoinableEvents([]);
    }
  }, [hostPlayer]);

  return {
    events,
    friendsEvents,
    groupJoinableEvents,
    updateInvite,
    cancelCommitment,
    joinTeeTime,
    refreshEvents,
  };
}
