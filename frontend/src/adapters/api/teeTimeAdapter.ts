import type { Event } from '../../domain/teeTime/types';
import type { TeeTimePort } from '../../ports/teeTimePort';
import { endpoints, request, authHeaders } from './httpClient';

export const teeTimeAdapter: TeeTimePort = {
  getEvents(playerId: number): Promise<Event[]> {
    return request<Event[]>(`${endpoints.players}/${playerId}/events`);
  },

  getFriendsEvents(playerId: number): Promise<Event[]> {
    return request<Event[]>(`${endpoints.players}/${playerId}/friends-events`);
  },

  createEvent(
    courseId: string,
    date: string,
    teeTime: string,
    openSpots: string,
    numHoles: string,
    isPrivate: boolean,
    hostId: number,
    selectedFriends: number[],
  ): Promise<Event> | undefined {
    if (!courseId || !teeTime) return undefined;
    return request<Event>(endpoints.singleEvent, {
      method: 'POST',
      body: JSON.stringify({
        course_id: courseId,
        date,
        tee_time: teeTime,
        open_spots: openSpots,
        number_of_holes: numHoles,
        private: isPrivate,
        host_id: hostId,
        invitees: selectedFriends,
      }),
    });
  },

  updateEvent(
    eventId: number,
    courseId: number,
    date: string,
    teeTime: string,
    openSpots: string,
    numHoles: string,
    isPrivate: boolean,
    invitees: number[] = [],
  ): Promise<Event> {
    return request<Event>(`${endpoints.singleEvent}/${eventId}`, {
      method: 'PATCH',
      body: JSON.stringify({
        course_id: String(courseId),
        date,
        tee_time: teeTime,
        open_spots: openSpots,
        number_of_holes: numHoles,
        private: isPrivate,
        invitees,
      }),
    });
  },

  deleteEvent(eventId: number, playerId: number): Promise<Event[]> {
    return fetch(`${endpoints.singleEvent}/${eventId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    }).then(() => this.getEvents(playerId));
  },

  updateInvite(playerId: number, eventId: number, status: string): Promise<Event[]> {
    return fetch(endpoints.playerEvent, {
      method: 'PATCH',
      body: JSON.stringify({
        player_id: playerId,
        event_id: eventId,
        invite_status: status,
      }),
      headers: authHeaders(),
    }).then(() => this.getEvents(playerId));
  },

  joinEvent(playerId: number, eventId: number): Promise<Event[]> {
    return fetch(endpoints.joinEvent, {
      method: 'POST',
      body: JSON.stringify({
        player_id: playerId,
        event_id: eventId,
      }),
      headers: authHeaders(),
    }).then(() => this.getEvents(playerId));
  },
};
