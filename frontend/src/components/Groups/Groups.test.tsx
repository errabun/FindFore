import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Route, Routes } from 'react-router-dom';
import GroupsPage from './GroupsPage';
import GroupDetailPage from './GroupDetailPage';
import { renderWithProviders } from '../../test/render';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import type { GroupSummary } from '../../domain/group/types';

vi.mock('../../adapters/api/groupAdapter', () => ({
  groupAdapter: {
    listMine: vi.fn(),
    listDiscover: vi.fn(),
    listInvitations: vi.fn(),
    get: vi.fn(),
    listMembers: vi.fn(),
    listJoinRequests: vi.fn(),
    listGroupInvitations: vi.fn(),
    join: vi.fn(),
    acceptInvitation: vi.fn(),
    declineInvitation: vi.fn(),
    listPosts: vi.fn(),
    createPost: vi.fn(),
    listEvents: vi.fn(),
  },
}));

const mocked = vi.mocked(groupAdapter);

const publicGroup: GroupSummary = {
  id: 10,
  name: 'Saturday Morning Golf',
  description: 'Early birds',
  privacy: 'public',
  owner: { id: 1, name: 'Eric' },
  member_count: 3,
  viewer_membership: null,
};

beforeEach(() => {
  vi.clearAllMocks();
  mocked.listMine.mockResolvedValue([]);
  mocked.listDiscover.mockResolvedValue([]);
  mocked.listInvitations.mockResolvedValue([]);
  mocked.listMembers.mockResolvedValue([]);
  mocked.listJoinRequests.mockResolvedValue([]);
  mocked.listGroupInvitations.mockResolvedValue([]);
  mocked.listPosts.mockResolvedValue([]);
  mocked.listEvents.mockResolvedValue([]);
});

describe('GroupsPage', () => {
  it('shows an empty state when the player has no groups', async () => {
    renderWithProviders(<GroupsPage hostPlayer={1} />);
    expect(await screen.findByText(/no groups yet/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /create group/i })).toHaveAttribute('href', '/groups/new');
  });

  it('accepts an outstanding invitation', async () => {
    const user = userEvent.setup();
    mocked.listInvitations.mockResolvedValue([
      {
        id: 7,
        group_id: 10,
        group_name: 'Saturday Morning Golf',
        inviter_player_id: 2,
        inviter_name: 'Sam',
        invitee_player_id: 1,
      },
    ]);
    mocked.acceptInvitation.mockResolvedValue({
      player_id: 1,
      player_name: 'Eric',
      role: 'member',
      status: 'active',
    });

    renderWithProviders(<GroupsPage hostPlayer={1} />);
    expect(await screen.findByText(/sam invited you to saturday morning golf/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /accept/i }));
    await waitFor(() => expect(mocked.acceptInvitation).toHaveBeenCalledWith(7));
  });
});

describe('GroupDetailPage', () => {
  it('lets a visitor join a public group', async () => {
    const user = userEvent.setup();
    mocked.get.mockResolvedValue(publicGroup);
    mocked.join.mockResolvedValue({
      player_id: 2,
      player_name: 'Sam',
      role: 'member',
      status: 'active',
    });

    renderWithProviders(
      <Routes>
        <Route path='/groups/:groupId' element={<GroupDetailPage hostPlayer={2} friends={[]} currentUserName='Sam' />} />
      </Routes>,
      { route: '/groups/10' },
    );

    expect(await screen.findByRole('button', { name: /join group/i })).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /join group/i }));
    await waitFor(() => expect(mocked.join).toHaveBeenCalledWith(10));
  });

  it('shows settings for the owner', async () => {
    mocked.get.mockResolvedValue({
      ...publicGroup,
      viewer_membership: { status: 'active', role: 'owner' },
    });
    mocked.listMembers.mockResolvedValue([
      { player_id: 1, player_name: 'Eric', role: 'owner', status: 'active' },
    ]);

    renderWithProviders(
      <Routes>
        <Route
          path='/groups/:groupId'
          element={<GroupDetailPage hostPlayer={1} friends={[{ id: 2, name: 'Sam' }]} currentUserName='Eric' />}
        />
      </Routes>,
      { route: '/groups/10' },
    );

    expect(await screen.findByRole('tab', { name: /settings/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /members/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /activity/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /rounds/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/share with the group/i)).toBeInTheDocument();
    expect(screen.getByText(/no posts yet/i)).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole('tab', { name: /rounds/i }));
    expect(await screen.findByText(/no upcoming rounds/i)).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /plan a round/i })[0]).toHaveAttribute(
      'href',
      '/event-form?group=10&name=Saturday%20Morning%20Golf',
    );
  });
});
