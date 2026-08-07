import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Login from './Login';
import { renderWithProviders } from '../../test/render';

describe('Login', () => {
  it('shows login error from props', () => {
    renderWithProviders(
      <Login validateLogin={vi.fn()} loginError='Too many login attempts. Please try again later.' />,
    );
    expect(screen.getByText(/too many login attempts/i)).toBeInTheDocument();
  });

  it('submits email and password to validateLogin', async () => {
    const user = userEvent.setup();
    const validateLogin = vi.fn();
    renderWithProviders(<Login validateLogin={validateLogin} />);

    await user.type(screen.getByLabelText(/email or username/i), 'eric@example.com');
    await user.type(document.getElementById('password') as HTMLElement, 'password1');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    expect(validateLogin).toHaveBeenCalledWith('eric@example.com', 'password1');
  });

  it('clears error when the user edits a field', async () => {
    const user = userEvent.setup();
    const clearLoginError = vi.fn();
    renderWithProviders(
      <Login validateLogin={vi.fn()} loginError='Invalid credentials' clearLoginError={clearLoginError} />,
    );

    await user.type(screen.getByLabelText(/email or username/i), 'x');
    expect(clearLoginError).toHaveBeenCalled();
  });
});
