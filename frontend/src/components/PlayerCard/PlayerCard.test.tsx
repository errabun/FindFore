import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import PlayerCard, { FriendRequestCard } from './PlayerCard';
import { renderWithProviders } from '../../test/render';
import type { HandleFriends } from '../../types';

const handlers: HandleFriends = {
  request: vi.fn(),
  remove: vi.fn(),
  accept: vi.fn(),
  decline: vi.fn(),
};

describe('PlayerCard relationship actions', () => {
  it('requests friendship when none exists', async () => {
    const user = userEvent.setup();
    const request = vi.fn();
    renderWithProviders(
      <PlayerCard
        playerInfo={{ id: 2, name: 'Ashley' }}
        friends={[]}
        incomingRequests={[]}
        outgoingPendingIds={[]}
        handleFriends={{ ...handlers, request }}
      />,
    );

    await user.click(screen.getByTitle(/request friend/i));
    expect(request).toHaveBeenCalledWith({ id: 2, name: 'Ashley' });
  });

  it('accepts and declines incoming requests', async () => {
    const user = userEvent.setup();
    const accept = vi.fn();
    const decline = vi.fn();
    renderWithProviders(
      <PlayerCard
        playerInfo={{ id: 2, name: 'Ashley' }}
        friends={[]}
        incomingRequests={[
          {
            id: 99,
            requesterId: 2,
            addresseeId: 1,
            status: 'pending',
            requesterName: 'Ashley',
            addresseeName: 'Eric',
          },
        ]}
        outgoingPendingIds={[]}
        handleFriends={{ ...handlers, accept, decline }}
      />,
    );

    await user.click(screen.getByTitle(/accept request/i));
    expect(accept).toHaveBeenCalledWith(99);

    await user.click(screen.getByTitle(/decline request/i));
    expect(decline).toHaveBeenCalledWith(99);
  });

  it('shows pending state for outgoing requests', () => {
    renderWithProviders(
      <PlayerCard
        playerInfo={{ id: 2, name: 'Ashley' }}
        friends={[]}
        incomingRequests={[]}
        outgoingPendingIds={[2]}
        handleFriends={handlers}
      />,
    );
    expect(screen.getByTitle(/request pending/i)).toBeDisabled();
  });
});

describe('FriendRequestCard', () => {
  it('wires Accept and Decline buttons', async () => {
    const user = userEvent.setup();
    const onAccept = vi.fn();
    const onDecline = vi.fn();
    renderWithProviders(
      <FriendRequestCard
        request={{
          id: 7,
          requesterId: 3,
          addresseeId: 1,
          status: 'pending',
          requesterName: 'Jordan',
          addresseeName: 'Eric',
        }}
        onAccept={onAccept}
        onDecline={onDecline}
      />,
    );

    await user.click(screen.getByRole('button', { name: /accept/i }));
    expect(onAccept).toHaveBeenCalledWith(7);
    await user.click(screen.getByRole('button', { name: /decline/i }));
    expect(onDecline).toHaveBeenCalledWith(7);
  });
});
