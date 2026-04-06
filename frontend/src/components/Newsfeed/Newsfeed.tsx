import { useState } from 'react';
import { Paper, Title, Stack, Badge, Group, Textarea, Button, Text } from '@mantine/core';
import { FiSend } from 'react-icons/fi';
import PostCard from './PostCard';
import { useNewsfeed } from '../../hooks/useNewsfeed';

interface NewsfeedProps {
  currentUserId: number;
  currentUserName: string;
}

const Newsfeed = ({ currentUserId, currentUserName }: NewsfeedProps) => {
  const { posts, createPost, deletePost, toggleReaction, createReply, deleteReply } = useNewsfeed(currentUserId);
  const [newPostBody, setNewPostBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleCreatePost = () => {
    if (!newPostBody.trim() || submitting) return;
    setSubmitting(true);
    createPost(newPostBody)
      ?.then(() => setNewPostBody(''))
      .finally(() => setSubmitting(false));
  };

  return (
    <Paper
      shadow='sm'
      style={{
        maxHeight: 'calc(100vh - 280px)',
        minHeight: 300,
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
        border: '1px solid var(--ff-border)',
      }}
    >
      <Group
        justify='space-between'
        align='center'
        px='md'
        py='sm'
        style={{ borderBottom: '1px solid var(--ff-border)' }}
      >
        <Group gap='sm'>
          <Title order={4} fw={600} style={{ color: 'var(--ff-heading)' }}>
            Community Feed
          </Title>
          <Badge size='sm' variant='light' color='forest'>
            {posts.length}
          </Badge>
        </Group>
      </Group>

      <Stack gap='md' p='md' style={{ overflowY: 'auto', flex: 1 }}>
        <Paper p='sm' withBorder style={{ borderColor: 'var(--ff-border)' }}>
          <Text size='xs' fw={500} style={{ color: 'var(--ff-label)' }} mb={4}>{currentUserName}</Text>
          <Textarea
            placeholder="What's on your mind?"
            value={newPostBody}
            onChange={(e) => setNewPostBody(e.target.value)}
            minRows={2}
            maxRows={4}
            autosize
            mb='xs'
          />
          <Group justify='flex-end'>
            <Button
              color='forest'
              size='xs'
              leftSection={<FiSend size={14} />}
              onClick={handleCreatePost}
              disabled={!newPostBody.trim() || submitting}
              loading={submitting}
            >
              Post
            </Button>
          </Group>
        </Paper>

        {posts.map((post) => (
          <PostCard
            key={post.id}
            post={post}
            currentUserId={currentUserId}
            onToggleReaction={toggleReaction}
            onCreateReply={createReply}
            onDeletePost={deletePost}
            onDeleteReply={deleteReply}
          />
        ))}

        {posts.length === 0 && (
          <Stack align='center' gap='xs' py='xl'>
            <Text size='sm' fw={600}>No posts yet</Text>
            <Text size='sm' c='dimmed'>Be the first to share something with the community!</Text>
          </Stack>
        )}
      </Stack>
    </Paper>
  );
};

export default Newsfeed;
