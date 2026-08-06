// Domain type re-exports (components keep importing from here)
export type { Player, LoginResponse, UpdateProfileRequest, ChangePasswordRequest } from './domain/auth/types';
export type { Event } from './domain/teeTime/types';
export type { Course } from './domain/course/types';
export type { Friend, FriendRequest, Post, Reaction, Reply } from './domain/social/types';

// UI-specific handler interfaces (component prop contracts, not domain)
export interface HandleFriends {
  request: (player: import('./domain/social/types').Friend | import('./domain/auth/types').Player) => void;
  remove: (friend: import('./domain/social/types').Friend) => void;
  accept: (requestId: number) => void;
  decline: (requestId: number) => void;
}

export interface HandleInviteAction {
  update: (eventId: number, status: string) => void;
  cancel: (event: import('./domain/teeTime/types').Event) => void;
  join: (eventId: number) => void;
}
