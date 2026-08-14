import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Dashboard from './Dashboard';
import { renderWithProviders } from '../../test/render';
import type { Event } from '../../types';

const joinable: Event = {
  id: 20,
  course_name: 'City Park',
  date: '2099-01-01',
  tee_time: '08:00',
  open_spots: 4,
  number_of_holes: '18',
  private: true,
  host_name: 'Eric',
  host_id: 1,
  accepted: [1],
  declined: [],
  pending: [],
  closed: [],
  remaining_spots: 3,
  group_id: 10,
  group_name: 'Saturday Morning Golf',
};

const baseProps = {
  events: [] as Event[],
  friendsEvents: [] as Event[],
  currentUserId: 2,
  currentUserName: 'Sam',
  screenWidth: 500,
  handleInviteAction: { update: vi.fn(), cancel: vi.fn(), join: vi.fn() },
  players: [],
  friends: [],
  incomingRequests: [],
  outgoingPendingIds: [],
  handleFriends: {
    request: vi.fn(),
    remove: vi.fn(),
    accept: vi.fn(),
    decline: vi.fn(),
  },
};

describe('Dashboard group need-one-more', () => {
  it('hides the groups card when nothing is joinable', () => {
    renderWithProviders(<Dashboard {...baseProps} />);
    expect(screen.queryByText(/groups need a player/i)).not.toBeInTheDocument();
  });

  it('shows joinable group rounds and lets the player join', async () => {
    const user = userEvent.setup();
    const join = vi.fn();
    renderWithProviders(
      <Dashboard {...baseProps} groupJoinableEvents={[joinable]} handleInviteAction={{ ...baseProps.handleInviteAction, join }} />,
    );

    expect(screen.getByText(/groups need a player/i)).toBeInTheDocument();
    expect(screen.getByText('City Park')).toBeInTheDocument();
    expect(screen.getByText('Saturday Morning Golf')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /join/i }));
    expect(join).toHaveBeenCalledWith(20);
  });
});
