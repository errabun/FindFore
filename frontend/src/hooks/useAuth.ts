import { useState, useEffect, useCallback } from 'react';
import type { Player, UpdateProfileRequest, ChangePasswordRequest } from '../domain/auth/types';
import { authAdapter } from '../adapters/api/authAdapter';
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
        // Refresh player graph so friend lists reflect follows made this session
        // before a full page reload (needed for Available → Friends filtering).
        try {
          const players = await authAdapter.getAllPlayers();
          setAllPlayers(players);
        } catch {
          // Keep stale list; dashboard still works with prior snapshot
        }
      })
      .catch(() => {
        setLoginError('Unable to sign in right now. Please try again.');
      });
  };

  const logout = useCallback(() => {
    clearSession();
    setHostPlayer(0);
  }, []);

  const clearLoginError = useCallback(() => setLoginError(''), []);

  // Load all players on mount
  useEffect(() => {
    authAdapter.getAllPlayers().then(setAllPlayers);
  }, []);

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
