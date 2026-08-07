import { useState, useEffect, useCallback } from 'react';
import type { Player, UpdateProfileRequest, ChangePasswordRequest } from '../domain/auth/types';
import { authAdapter } from '../adapters/api/authAdapter';
import { ApiError, setUnauthorizedHandler } from '../adapters/api/httpClient';
import {
  getPlayerIdFromToken,
  isSessionActive,
  touchActivity,
  setToken,
  clearSession,
} from '../adapters/storage/localStorageAdapter';

export function useAuth() {
  const restoredId = getPlayerIdFromToken();
  const initialPlayer = restoredId && isSessionActive() ? restoredId : 0;

  const [hostPlayer, setHostPlayer] = useState(initialPlayer);
  const [loginError, setLoginError] = useState('');
  const [allPlayers, setAllPlayers] = useState<Player[]>([]);

  const logout = useCallback(() => {
    clearSession();
    setHostPlayer(0);
    setAllPlayers([]);
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      logout();
    });
    return () => setUnauthorizedHandler(null);
  }, [logout]);

  const validateLogin = (email: string, password: string) => {
    setLoginError('');
    authAdapter
      .login(email, password)
      .then(async (data) => {
        if (!data) {
          setLoginError('Invalid email or password. Please try again.');
          return;
        }
        setHostPlayer(data.id);
        if (data.token) {
          setToken(data.token);
        }
        touchActivity();
        try {
          const players = await authAdapter.getAllPlayers();
          setAllPlayers(players);
        } catch {
          // Keep empty list; dashboard still works for the signed-in user
        }
      })
      .catch((err) => {
        if (err instanceof ApiError) {
          setLoginError(err.message);
          return;
        }
        setLoginError('Unable to sign in right now. Please try again.');
      });
  };

  const clearLoginError = useCallback(() => setLoginError(''), []);

  // Load community players only when authenticated (endpoint requires JWT).
  useEffect(() => {
    if (!hostPlayer) {
      setAllPlayers([]);
      return;
    }
    authAdapter.getAllPlayers().then(setAllPlayers).catch(() => setAllPlayers([]));
  }, [hostPlayer]);

  // Track activity and enforce session timeout
  useEffect(() => {
    if (!hostPlayer) return;

    const onActivity = () => touchActivity();
    window.addEventListener('click', onActivity);
    window.addEventListener('keydown', onActivity);

    const interval = setInterval(() => {
      if (!isSessionActive()) {
        logout();
      }
    }, 60_000);

    return () => {
      window.removeEventListener('click', onActivity);
      window.removeEventListener('keydown', onActivity);
      clearInterval(interval);
    };
  }, [hostPlayer, logout]);

  const updateProfile = async (data: UpdateProfileRequest) => {
    await authAdapter.updateProfile(hostPlayer, data);
    const players = await authAdapter.getAllPlayers();
    setAllPlayers(players);
  };

  const changePassword = (data: ChangePasswordRequest) => {
    return authAdapter.changePassword(hostPlayer, data);
  };

  return {
    hostPlayer,
    loginError,
    allPlayers,
    validateLogin,
    logout,
    clearLoginError,
    updateProfile,
    changePassword,
  };
}
