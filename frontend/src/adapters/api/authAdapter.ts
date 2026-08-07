import type { Player, LoginResponse, UpdateProfileRequest, ChangePasswordRequest } from '../../domain/auth/types';
import type { AuthPort } from '../../ports/authPort';
import { endpoints, request, requestVoid, requestPublic, ApiError } from './httpClient';

export const authAdapter: AuthPort = {
  async login(login: string, password: string): Promise<LoginResponse | undefined> {
    try {
      return await requestPublic<LoginResponse>(endpoints.sessions, {
        method: 'POST',
        body: JSON.stringify({ login, password }),
      });
    } catch (err) {
      if (err instanceof ApiError && (err.status === 401 || err.status === 429)) {
        throw err;
      }
      return undefined;
    }
  },

  createProfile(
    name: string,
    phone: string,
    email: string,
    username: string,
    password: string,
    passwordConfirmation: string,
  ): Promise<Player> {
    return requestPublic<Player>(endpoints.players, {
      method: 'POST',
      body: JSON.stringify({
        name,
        phone,
        email,
        username,
        password,
        password_confirmation: passwordConfirmation,
      }),
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
