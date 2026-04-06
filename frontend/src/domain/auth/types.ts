export interface Player {
  id: number;
  name: string;
  phone: string;
  email: string;
  username: string;
  friends: number[];
  events: number[];
}

export interface LoginResponse {
  id: number;
  name: string;
  phone: string;
  email: string;
  username: string;
  friends: number[];
  events: number[];
  token: string;
}

export interface UpdateProfileRequest {
  name: string;
  phone: string;
  email: string;
  username: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  password_confirmation: string;
}
