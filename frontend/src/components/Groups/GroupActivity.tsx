import { useState } from 'react';
import { Alert, Button, Group, Skeleton, Stack, Text, Textarea } from '@mantine/core';
import { FiAlertCircle, FiMessageSquare, FiSend } from 'react-icons/fi';
import PostCard from '../Newsfeed/PostCard';
import EmptyState from '../EmptyState/EmptyState';
import { useGroupFeed } from '../../hooks/useGroupFeed';

interface GroupActivityProps {
  groupId: number;
  currentUserId: number;
  currentUserName: string;
  canManage: boolean;
}

export default function GroupActivity({
  groupId,
  currentUserId,
  currentUserName,
  canManage,
}: GroupActivityProps) {
  const { posts, loading, error, createPost, deletePost, toggleReaction, createReply, deleteReply } =
    useGroupFeed(groupId, currentUserId);
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const submit = () => {
    if (!body.trim() || submitting) return;
    setSubmitting(true);
    createPost(body)
      .then(() => setBody(''))
      .finally(() => setSubmitting(false));
  };

  return (
    <Stack gap='md'>
      {error && (
        <Alert color='red' icon={<FiAlertCircle />}>
          {error}
        </Alert>
      )}
      <div>
        <Text size='xs' fw={500} mb={4} style={{ color: 'var(--ff-label)' }}>
          {currentUserName}
        </Text>
        <Textarea
          label='Share with the group'
          placeholder="What's happening with the group?"
          value={body}
          onChange={(e) => setBody(e.currentTarget.value)}
          minRows={3}
          mb='xs'
        />
        <Group justify='flex-end'>
          <Button
            color='forest'
            size='md'
            leftSection={<FiSend size={14} />}
            onClick={submit}
            disabled={!body.trim() || submitting}
            loading={submitting}
          >
            Post
          </Button>
        </Group>
      </div>

      {loading && posts.length === 0 ? (
        <Stack gap='sm'>
          <Skeleton height={120} radius='md' />
          <Skeleton height={80} radius='md' />
        </Stack>
      ) : posts.length === 0 ? (
        <EmptyState
          icon={<FiMessageSquare />}
          title='No posts yet'
          description='Share an update, ask who is in, or recap the last round.'
        />
      ) : (
        posts.map((post) => (
          <PostCard
            key={post.id}
            post={post}
            currentUserId={currentUserId}
            canDelete={post.player_id === currentUserId || canManage}
            onToggleReaction={toggleReaction}
            onCreateReply={createReply}
            onDeletePost={deletePost}
            onDeleteReply={deleteReply}
          />
        ))
      )}
    </Stack>
  );
}
