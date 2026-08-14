import { useCallback, useEffect, useState } from 'react';
import { groupAdapter } from '../adapters/api/groupAdapter';
import { newsfeedAdapter } from '../adapters/api/socialAdapter';
import { ApiError } from '../adapters/api/httpClient';
import type { Post } from '../domain/social/types';

function messageFrom(err: unknown, fallback: string) {
  return err instanceof ApiError ? err.message : fallback;
}

export function useGroupFeed(groupId: number, hostPlayer: number) {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(Boolean(hostPlayer && groupId));
  const [error, setError] = useState('');

  const refresh = useCallback(() => {
    if (!hostPlayer || !groupId) {
      setPosts([]);
      return Promise.resolve();
    }
    setLoading(true);
    setError('');
    return groupAdapter
      .listPosts(groupId)
      .then(setPosts)
      .catch((err) => setError(messageFrom(err, 'Could not load group posts')))
      .finally(() => setLoading(false));
  }, [hostPlayer, groupId]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createPost = (body: string) => {
    if (!body.trim()) return Promise.resolve();
    return groupAdapter.createPost(groupId, body.trim()).then((post) => {
      setPosts((prev) => [post, ...prev]);
    });
  };

  const deletePost = (postId: number) => {
    newsfeedAdapter.deletePost(postId).then(() => {
      setPosts((prev) => prev.filter((p) => p.id !== postId));
    });
  };

  const toggleReaction = (postId: number, emoji: string) => {
    newsfeedAdapter.toggleReaction(postId, emoji).then((reactions) => {
      setPosts((prev) => prev.map((p) => (p.id === postId ? { ...p, reactions } : p)));
    });
  };

  const createReply = (postId: number, body: string) => {
    newsfeedAdapter.createReply(postId, body).then((reply) => {
      setPosts((prev) =>
        prev.map((p) => (p.id === postId ? { ...p, replies: [...p.replies, reply] } : p)),
      );
    });
  };

  const deleteReply = (postId: number, replyId: number) => {
    newsfeedAdapter.deleteReply(postId, replyId).then(() => {
      setPosts((prev) =>
        prev.map((p) =>
          p.id === postId ? { ...p, replies: p.replies.filter((r) => r.id !== replyId) } : p,
        ),
      );
    });
  };

  return { posts, loading, error, createPost, deletePost, toggleReaction, createReply, deleteReply };
}
