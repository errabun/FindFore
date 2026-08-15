import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import TeeTimePicker, { normalizeTeeTimeValue } from './TeeTimePicker';
import { renderWithProviders } from '../../test/render';

describe('normalizeTeeTimeValue', () => {
  it('strips seconds so TimePicker accepts API values', () => {
    expect(normalizeTeeTimeValue('08:00:00')).toBe('08:00');
    expect(normalizeTeeTimeValue('8:00')).toBe('08:00');
    expect(normalizeTeeTimeValue('')).toBe('');
  });
});

describe('TeeTimePicker', () => {
  it('labels the field and shows the 5am–8pm window', () => {
    renderWithProviders(<TeeTimePicker value='' onChange={vi.fn()} />);

    expect(screen.getByText('Tee time')).toBeInTheDocument();
    expect(screen.getByText(/between 5:00 am and 8:00 pm/i)).toBeInTheDocument();
    expect(screen.getByLabelText('Hours')).toBeInTheDocument();
  });

  it('shows the selected time in 12-hour fields', () => {
    renderWithProviders(<TeeTimePicker value='08:00' onChange={vi.fn()} />);

    expect(screen.getByLabelText('Hours')).toHaveValue('08');
    expect(screen.getByLabelText('Minutes')).toHaveValue('00');
    expect(screen.getByLabelText('AM/PM')).toHaveValue('AM');
  });
});
