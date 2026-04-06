import type { Post, Reaction, Reply } from '../../domain/social/types';
import type { FriendshipPort, FriendshipResponse, NewsfeedPort } from '../../ports/socialPort';
import { endpoints, request, requestVoid, requestRaw } from './httpClient';

export const friendshipAdapter: FriendshipPort = {
  follow(followerId: number, followeeId: number): Promise<FriendshipResponse> {
    return request<FriendshipResponse>(endpoints.friendship, {
      method: 'POST',
      body: JSON.stringify({
        follower_id: followerId,
        followee_id: followeeId,
      }),
    });
  },

  unfollow(followerId: number, followeeId: number): Promise<Response> {
    return requestRaw(endpoints.friendship, {
      method: 'DELETE',
      body: JSON.stringify({
        follower_id: followerId,
        followee_id: followeeId,
      }),
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
