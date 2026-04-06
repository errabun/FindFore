import { useState, useEffect } from 'react';
import type { Player } from '../domain/auth/types';
import type { Friend } from '../domain/social/types';
import { buildFriendList } from '../domain/auth/authService';
import { friendshipAdapter } from '../adapters/api/socialAdapter';

export function useFriends(hostPlayer: number, allPlayers: Player[]) {
  const [friends, setFriends] = useState<Friend[]>([]);

  const addFriend = (friend: Friend) => {
    if (!hostPlayer) return;
    friendshipAdapter.follow(hostPlayer, friend.id).then((data) => {
      setFriends((prev) => [
        ...prev,
        { id: data.followee.id, name: data.followee.name },
      ]);
    });
  };

  const removeFriend = (unFriend: Friend) => {
    if (!hostPlayer) return;
    friendshipAdapter.unfollow(hostPlayer, unFriend.id).then(() => {
      setFriends((prev) => prev.filter((f) => f.id !== unFriend.id));
    });
  };

  // Derive friend list when hostPlayer or allPlayers change
  useEffect(() => {
    if (!hostPlayer) {
      setFriends([]);
      return;
    }
    const currentPlayer = allPlayers.find((p) => p.id === hostPlayer);
    if (currentPlayer?.friends) {
      setFriends(buildFriendList(allPlayers, currentPlayer));
    }
  }, [allPlayers, hostPlayer]);

  return { friends, addFriend, removeFriend };
}
