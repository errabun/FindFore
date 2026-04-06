import type { Player } from '../domain/auth/types';
import type { Post, Reaction, Reply } from '../domain/social/types';

export interface FriendshipResponse {
  id: number;
  follower_id: number;
  followee_id: number;
  follower: Player;
  followee: Player;
}

export interface FriendshipPort {
  follow(followerId: number, followeeId: number): Promise<FriendshipResponse>;
  unfollow(followerId: number, followeeId: number): Promise<Response>;
}

export interface NewsfeedPort {
  getPosts(limit?: number, offset?: number): Promise<Post[]>;
  createPost(playerId: number, body: string): Promise<Post>;
  deletePost(postId: number, playerId: number): Promise<void>;
  toggleReaction(postId: number, playerId: number, emoji: string): Promise<Reaction[]>;
  createReply(postId: number, playerId: number, body: string): Promise<Reply>;
  deleteReply(postId: number, replyId: number, playerId: number): Promise<void>;
}
