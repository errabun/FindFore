import { describe, it, expect } from 'vitest';
import {
  filterAvailable,
  filterFriendInvites,
  filterPublicInvites,
  filterCommitted,
} from './teeTimeService';
import type { Event } from './types';

function makeEvent(overrides: Partial<Event>): Event {
  return {
    id: 1,
    course_name: 'Test Course',
    date: '2099-01-01',
    tee_time: '08:00',
    open_spots: 2,
    number_of_holes: '18',
    private: false,
    host_name: 'Host',
    host_id: 10,
    accepted: [],
    declined: [],
    pending: [],
    closed: [],
    remaining_spots: 2,
    ...overrides,
  };
}

describe('teeTimeService filters', () => {
  const me = 1;

  it('filterAvailable keeps public and pending invites, drops accepted/declined/closed', () => {
    const events = [
      makeEvent({ id: 1, private: false }),
      makeEvent({ id: 2, private: true, pending: [me] }),
      makeEvent({ id: 3, private: true, pending: [99] }),
      makeEvent({ id: 4, private: false, accepted: [me] }),
      makeEvent({ id: 5, private: false, declined: [me] }),
      makeEvent({ id: 6, private: false, closed: [me] }),
    ];

    const available = filterAvailable(events, me).map((e) => e.id);
    expect(available).toEqual([1, 2]);
  });

  it('filterPublicInvites keeps only non-private events', () => {
    const events = [
      makeEvent({ id: 1, private: false }),
      makeEvent({ id: 2, private: true }),
    ];
    expect(filterPublicInvites(events).map((e) => e.id)).toEqual([1]);
  });

  it('filterFriendInvites keeps events hosted by friends', () => {
    const events = [
      makeEvent({ id: 1, host_id: 10 }),
      makeEvent({ id: 2, host_id: 20 }),
    ];
    expect(filterFriendInvites(events, [10]).map((e) => e.id)).toEqual([1]);
  });

  it('filterCommitted keeps accepted events for the player', () => {
    const events = [
      makeEvent({ id: 1, accepted: [me] }),
      makeEvent({ id: 2, accepted: [99] }),
    ];
    expect(filterCommitted(events, me).map((e) => e.id)).toEqual([1]);
  });
});
