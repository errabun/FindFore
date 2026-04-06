import { useState, useEffect, useCallback } from 'react';
import type { Post } from '../domain/social/types';
import { newsfeedAdapter } from '../adapters/api/socialAdapter';

export function useNewsfeed(currentUserId: number) {
  const [posts, setPosts] = useState<Post[]>([]);

  const fetchPosts = useCallback(() => {
    newsfeedAdapter.getPosts().then(setPosts);
  }, []);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  const createPost = (body: string) => {
    if (!body.trim()) return Promise.resolve();
    return newsfeedAdapter.createPost(currentUserId, body.trim()).then((post) => {
      setPosts((prev) => [post, ...prev]);
    });
  };

  const deletePost = (postId: number) => {
    newsfeedAdapter.deletePost(postId, currentUserId).then(() => {
      setPosts((prev) => prev.filter((p) => p.id !== postId));
    });
  };

  const toggleReaction = (postId: number, emoji: string) => {
    newsfeedAdapter.toggleReaction(postId, currentUserId, emoji).then((reactions) => {
      setPosts((prev) =>
        prev.map((p) => (p.id === postId ? { ...p, reactions } : p)),
      );
    });
  };

  const createReply = (postId: number, body: string) => {
    newsfeedAdapter.createReply(postId, currentUserId, body).then((reply) => {
      setPosts((prev) =>
        prev.map((p) =>
          p.id === postId ? { ...p, replies: [...p.replies, reply] } : p,
        ),
      );
    });
  };

  const deleteReply = (postId: number, replyId: number) => {
    newsfeedAdapter.deleteReply(postId, replyId, currentUserId).then(() => {
      setPosts((prev) =>
        prev.map((p) =>
          p.id === postId
            ? { ...p, replies: p.replies.filter((r) => r.id !== replyId) }
            : p,
        ),
      );
    });
  };

  return {
    posts,
    createPost,
    deletePost,
    toggleReaction,
    createReply,
    deleteReply,
  };
}
