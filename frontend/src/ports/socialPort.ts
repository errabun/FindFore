import type { Player } from '../domain/auth/types';
import type { FriendRequest, Post, Reaction, Reply } from '../domain/social/types';

export interface FriendshipResponse {
  id: number;
  requester_id: number;
  addressee_id: number;
  status: string;
  requester: Player;
  addressee: Player;
}

export interface FriendshipPort {
  listAccepted(): Promise<FriendshipResponse[]>;
  listIncomingRequests(): Promise<FriendshipResponse[]>;
  listOutgoingPendingIds(): Promise<number[]>;
  request(playerId: number): Promise<FriendshipResponse>;
  accept(friendshipId: number): Promise<FriendshipResponse>;
  decline(friendshipId: number): Promise<void>;
  remove(friendshipId: number): Promise<void>;
}

export interface NewsfeedPort {
  getPosts(limit?: number, offset?: number): Promise<Post[]>;
  createPost(body: string): Promise<Post>;
  deletePost(postId: number): Promise<void>;
  toggleReaction(postId: number, emoji: string): Promise<Reaction[]>;
  createReply(postId: number, body: string): Promise<Reply>;
  deleteReply(postId: number, replyId: number): Promise<void>;
}

export function mapIncomingRequest(row: FriendshipResponse): FriendRequest {
  return {
    id: row.id,
    requesterId: row.requester_id,
    addresseeId: row.addressee_id,
    status: row.status,
    requesterName: row.requester.name,
    addresseeName: row.addressee.name,
  };
}
