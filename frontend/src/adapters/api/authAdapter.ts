import type { Player, LoginResponse, UpdateProfileRequest, ChangePasswordRequest } from '../../domain/auth/types';
import type { AuthPort } from '../../ports/authPort';
import { endpoints, request, requestVoid } from './httpClient';

export const authAdapter: AuthPort = {
  login(login: string, password: string): Promise<LoginResponse | undefined> {
    return fetch(endpoints.sessions, {
      method: 'POST',
      body: JSON.stringify({ login, password }),
      headers: { 'Content-Type': 'application/json' },
    }).then(resp => {
      if (resp.ok) return resp.json();
      return undefined;
    });
  },

  createProfile(
    name: string,
    phone: string,
    email: string,
    username: string,
    password: string,
    passwordConfirmation: string,
  ): Promise<Player> {
    return fetch(endpoints.players, {
      method: 'POST',
      body: JSON.stringify({
        name,
        phone,
        email,
        username,
        password,
        password_confirmation: passwordConfirmation,
      }),
      headers: { 'Content-Type': 'application/json' },
    }).then(resp => {
      if (resp.ok) return resp.json();
      throw new Error('Unable to create new profile, please try again!');
    });
  },

  getAllPlayers(): Promise<Player[]> {
    return request<Player[]>(endpoints.players);
  },

  updateProfile(playerId: number, data: UpdateProfileRequest): Promise<Player> {
    return request<Player>(`${endpoints.players}/${playerId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  changePassword(playerId: number, data: ChangePasswordRequest): Promise<void> {
    return requestVoid(`${endpoints.players}/${playerId}/password`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },
};
