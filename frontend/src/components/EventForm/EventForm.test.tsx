import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EventForm from './EventForm';
import { renderWithProviders } from '../../test/render';

vi.mock('../../adapters/api/courseAdapter', () => ({
  courseAdapter: {
    search: vi.fn().mockResolvedValue([]),
    findOrCreate: vi.fn(),
  },
}));

vi.mock('../../adapters/api/teeTimeAdapter', () => ({
  teeTimeAdapter: {
    createEvent: vi.fn(),
  },
}));

describe('EventForm privacy UI', () => {
  it('hides invite friends until Private is selected', async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <EventForm courses={[]} friends={[{ id: 2, name: 'Ashley' }]} hostId={1} refreshEvents={vi.fn()} />,
    );

    expect(screen.queryByText(/invite friends/i)).not.toBeInTheDocument();

    await user.click(screen.getByLabelText(/^private$/i));
    expect(screen.getByText(/invite friends/i)).toBeInTheDocument();
    expect(screen.getByText('Ashley')).toBeInTheDocument();
  });

  it('hides public/private and friend invites for a group round', () => {
    renderWithProviders(
      <EventForm courses={[]} friends={[{ id: 2, name: 'Ashley' }]} hostId={1} refreshEvents={vi.fn()} />,
      { route: '/event-form?group=10&name=Saturday+Morning+Golf' },
    );

    expect(screen.getByText(/plan a group round/i)).toBeInTheDocument();
    expect(screen.getByText(/this round stays in saturday morning golf/i)).toBeInTheDocument();
    expect(screen.queryByText(/public or private/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/invite friends/i)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /plan round/i })).toBeInTheDocument();
  });
});
