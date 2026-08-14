import type { Event } from '../domain/teeTime/types';

export interface TeeTimePort {
  getEvents(playerId: number): Promise<Event[]>;
  getFriendsEvents(playerId: number): Promise<Event[]>;
  getGroupJoinableEvents(): Promise<Event[]>;
  createEvent(
    courseId: string,
    date: string,
    teeTime: string,
    openSpots: string,
    numHoles: string,
    isPrivate: boolean,
    hostId: number,
    selectedFriends: number[],
    groupId?: number,
  ): Promise<Event> | undefined;
  updateEvent(
    eventId: number,
    courseId: number,
    date: string,
    teeTime: string,
    openSpots: string,
    numHoles: string,
    isPrivate: boolean,
    invitees?: number[],
  ): Promise<Event>;
  deleteEvent(eventId: number, playerId: number): Promise<Event[]>;
  updateInvite(playerId: number, eventId: number, status: string): Promise<Event[]>;
  joinEvent(playerId: number, eventId: number): Promise<Event[]>;
}
