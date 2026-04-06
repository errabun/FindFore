import type { Player, LoginResponse, UpdateProfileRequest, ChangePasswordRequest } from '../domain/auth/types';

export interface AuthPort {
  login(login: string, password: string): Promise<LoginResponse | undefined>;
  createProfile(
    name: string,
    phone: string,
    email: string,
    username: string,
    password: string,
    passwordConfirmation: string,
  ): Promise<Player>;
  getAllPlayers(): Promise<Player[]>;
  updateProfile(playerId: number, data: UpdateProfileRequest): Promise<Player>;
  changePassword(playerId: number, data: ChangePasswordRequest): Promise<void>;
}
