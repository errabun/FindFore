import { useState, useEffect, useCallback } from 'react';
import type { Event } from '../domain/teeTime/types';
import { teeTimeAdapter } from '../adapters/api/teeTimeAdapter';

export function useTeeTimes(hostPlayer: number) {
  const [events, setEvents] = useState<Event[]>([]);
  const [friendsEvents, setFriendsEvents] = useState<Event[]>([]);

  const refreshEvents = useCallback(() => {
    if (!hostPlayer) return;
    teeTimeAdapter.getEvents(hostPlayer).then(setEvents);
    teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
  }, [hostPlayer]);

  const updateInvite = (eventId: number, status: string) => {
    teeTimeAdapter.updateInvite(hostPlayer, eventId, status).then((events) => {
      setEvents(events);
      teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
    });
  };

  const cancelCommitment = (event: Event) => {
    if (event.host_id === hostPlayer) {
      teeTimeAdapter.deleteEvent(event.id, hostPlayer).then((events) => {
        setEvents(events);
        teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
      });
    } else {
      teeTimeAdapter.updateInvite(hostPlayer, event.id, 'declined').then((events) => {
        setEvents(events);
        teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
      });
    }
  };

  const joinTeeTime = (eventId: number) => {
    teeTimeAdapter.joinEvent(hostPlayer, eventId).then((events) => {
      setEvents(events);
      setFriendsEvents((prev) => prev.filter((e) => e.id !== eventId));
    });
  };

  // Load events when hostPlayer changes
  useEffect(() => {
    if (hostPlayer) {
      teeTimeAdapter.getEvents(hostPlayer).then(setEvents);
      teeTimeAdapter.getFriendsEvents(hostPlayer).then(setFriendsEvents);
    } else {
      setEvents([]);
      setFriendsEvents([]);
    }
  }, [hostPlayer]);

  return {
    events,
    friendsEvents,
    updateInvite,
    cancelCommitment,
    joinTeeTime,
    refreshEvents,
  };
}
