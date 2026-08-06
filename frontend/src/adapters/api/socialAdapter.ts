import type { Post, Reaction, Reply } from '../../domain/social/types';
import type { FriendshipPort, FriendshipResponse, NewsfeedPort } from '../../ports/socialPort';
import { endpoints, request, requestVoid } from './httpClient';

export const friendshipAdapter: FriendshipPort = {
  listAccepted(): Promise<FriendshipResponse[]> {
    return request<FriendshipResponse[]>(endpoints.friendships);
  },

  listIncomingRequests(): Promise<FriendshipResponse[]> {
    return request<FriendshipResponse[]>(`${endpoints.friendships}/requests`);
  },

  listOutgoingPendingIds(): Promise<number[]> {
    return request<number[]>(`${endpoints.friendships}/outgoing`);
  },

  request(playerId: number): Promise<FriendshipResponse> {
    return request<FriendshipResponse>(endpoints.friendships, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId }),
    });
  },

  accept(friendshipId: number): Promise<FriendshipResponse> {
    return request<FriendshipResponse>(`${endpoints.friendships}/${friendshipId}/accept`, {
      method: 'POST',
    });
  },

  decline(friendshipId: number): Promise<void> {
    return requestVoid(`${endpoints.friendships}/${friendshipId}/decline`, {
      method: 'POST',
    });
  },

  remove(friendshipId: number): Promise<void> {
    return requestVoid(`${endpoints.friendships}/${friendshipId}`, {
      method: 'DELETE',
    });
  },
};

export const newsfeedAdapter: NewsfeedPort = {
  getPosts(limit = 50, offset = 0): Promise<Post[]> {
    return request<Post[]>(`${endpoints.posts}?limit=${limit}&offset=${offset}`);
  },

  createPost(playerId: number, body: string): Promise<Post> {
    return request<Post>(endpoints.posts, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId, body }),
    });
  },

  deletePost(postId: number, playerId: number): Promise<void> {
    return requestVoid(`${endpoints.posts}/${postId}`, {
      method: 'DELETE',
      body: JSON.stringify({ player_id: playerId }),
    });
  },

  toggleReaction(postId: number, playerId: number, emoji: string): Promise<Reaction[]> {
    return request<Reaction[]>(`${endpoints.posts}/${postId}/reactions`, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId, emoji }),
    });
  },

  createReply(postId: number, playerId: number, body: string): Promise<Reply> {
    return request<Reply>(`${endpoints.posts}/${postId}/replies`, {
      method: 'POST',
      body: JSON.stringify({ player_id: playerId, body }),
    });
  },

  deleteReply(postId: number, replyId: number, playerId: number): Promise<void> {
    return requestVoid(`${endpoints.posts}/${postId}/replies/${replyId}`, {
      method: 'DELETE',
      body: JSON.stringify({ player_id: playerId }),
    });
  },
};
