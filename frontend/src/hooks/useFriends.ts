import { useState, useEffect, useCallback } from 'react';
import type { Player } from '../domain/auth/types';
import type { Friend, FriendRequest } from '../domain/social/types';
import { friendshipAdapter } from '../adapters/api/socialAdapter';
import { mapIncomingRequest, type FriendshipResponse } from '../ports/socialPort';

function friendFromAccepted(row: FriendshipResponse, viewerId: number): Friend {
  const other = row.requester_id === viewerId ? row.addressee : row.requester;
  return {
    id: other.id,
    name: other.name,
    friendshipId: row.id,
  };
}

export function useFriends(hostPlayer: number, allPlayers: Player[]) {
  const [friends, setFriends] = useState<Friend[]>([]);
  const [incomingRequests, setIncomingRequests] = useState<FriendRequest[]>([]);
  const [outgoingPendingIds, setOutgoingPendingIds] = useState<number[]>([]);

  const refreshFriendships = useCallback(() => {
    if (!hostPlayer) {
      setFriends([]);
      setIncomingRequests([]);
      setOutgoingPendingIds([]);
      return Promise.resolve();
    }

    return Promise.all([
      friendshipAdapter.listAccepted(),
      friendshipAdapter.listIncomingRequests(),
      friendshipAdapter.listOutgoingPendingIds(),
    ]).then(([accepted, incoming, outgoing]) => {
      setFriends(accepted.map((row) => friendFromAccepted(row, hostPlayer)));
      setIncomingRequests(incoming.map(mapIncomingRequest));
      setOutgoingPendingIds(outgoing);
    }).catch(() => {
      // Keep prior snapshot on transient failures
    });
  }, [hostPlayer]);

  useEffect(() => {
    refreshFriendships();
  }, [refreshFriendships, allPlayers]);

  const requestFriend = (player: Friend | Player) => {
    if (!hostPlayer) return;
    friendshipAdapter.request(player.id).then((row) => {
      if (row.status === 'accepted') {
        setFriends((prev) => {
          if (prev.some((f) => f.id === player.id)) return prev;
          return [...prev, friendFromAccepted(row, hostPlayer)];
        });
        setIncomingRequests((prev) => prev.filter((r) => r.id !== row.id));
        setOutgoingPendingIds((prev) => prev.filter((id) => id !== player.id));
        return;
      }
      setOutgoingPendingIds((prev) =>
        prev.includes(player.id) ? prev : [...prev, player.id],
      );
    });
  };

  const acceptRequest = (requestId: number) => {
    if (!hostPlayer) return;
    friendshipAdapter.accept(requestId).then((row) => {
      setIncomingRequests((prev) => prev.filter((r) => r.id !== requestId));
      setFriends((prev) => {
        const friend = friendFromAccepted(row, hostPlayer);
        if (prev.some((f) => f.id === friend.id)) return prev;
        return [...prev, friend];
      });
    });
  };

  const declineRequest = (requestId: number) => {
    if (!hostPlayer) return;
    friendshipAdapter.decline(requestId).then(() => {
      setIncomingRequests((prev) => prev.filter((r) => r.id !== requestId));
    });
  };

  const removeFriend = (friend: Friend) => {
    if (!hostPlayer || !friend.friendshipId) return;
    friendshipAdapter.remove(friend.friendshipId).then(() => {
      setFriends((prev) => prev.filter((f) => f.id !== friend.id));
      setOutgoingPendingIds((prev) => prev.filter((id) => id !== friend.id));
    });
  };

  return {
    friends,
    incomingRequests,
    outgoingPendingIds,
    requestFriend,
    acceptRequest,
    declineRequest,
    removeFriend,
    refreshFriendships,
  };
}
