import { useState } from 'react';
import { Stack, Group, Text, TextInput, ActionIcon, UnstyledButton, Avatar } from '@mantine/core';
import { FiSend, FiTrash2 } from 'react-icons/fi';
import dayjs from 'dayjs';
import type { Reply } from '../../types';

interface ReplySectionProps {
  replies: Reply[];
  currentUserId: number;
  onCreateReply: (body: string) => void;
  onDeleteReply: (replyId: number) => void;
  defaultShowInput?: boolean;
}

const ReplySection = ({ replies, currentUserId, onCreateReply, onDeleteReply, defaultShowInput = false }: ReplySectionProps) => {
  const [replyText, setReplyText] = useState('');
  const [showAll, setShowAll] = useState(false);

  const handleSubmit = () => {
    if (!replyText.trim()) return;
    onCreateReply(replyText.trim());
    setReplyText('');
  };

  const visibleReplies = showAll ? replies : replies.slice(0, 2);
  const hiddenCount = replies.length - 2;

  return (
    <Stack gap='xs'>
      {visibleReplies.length > 0 && (
        <Stack gap='xs' pl='sm' style={{ borderLeft: '2px solid var(--ff-border)' }}>
          {visibleReplies.map((reply) => (
            <Group key={reply.id} gap='xs' align='flex-start' wrap='nowrap'>
              <Avatar size='xs' radius='xl' color='forest' mt={2}>
                {reply.player_name.charAt(0)}
              </Avatar>
              <div style={{ flex: 1, minWidth: 0 }}>
                <Group gap='xs'>
                  <Text size='xs' fw={600} style={{ color: 'var(--ff-label)' }}>{reply.player_name}</Text>
                  <Text size='xs' c='dimmed'>{dayjs(reply.created_at).format('MMM D, h:mm A')}</Text>
                </Group>
                <Text size='sm'>{reply.body}</Text>
              </div>
              {reply.player_id === currentUserId && (
                <ActionIcon variant='subtle' color='red' size='xs' onClick={() => onDeleteReply(reply.id)}>
                  <FiTrash2 size={12} />
                </ActionIcon>
              )}
            </Group>
          ))}

          {!showAll && hiddenCount > 0 && (
            <UnstyledButton onClick={() => setShowAll(true)}>
              <Text size='xs' style={{ color: 'var(--ff-link)' }} fw={500}>
                View {hiddenCount} more {hiddenCount === 1 ? 'reply' : 'replies'}
              </Text>
            </UnstyledButton>
          )}
        </Stack>
      )}

      {defaultShowInput && (
        <Group gap='xs'>
          <TextInput
            placeholder='Write a reply...'
            size='xs'
            value={replyText}
            onChange={(e) => setReplyText(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
            style={{ flex: 1 }}
          />
          <ActionIcon
            color='forest'
            variant='filled'
            size='sm'
            onClick={handleSubmit}
            disabled={!replyText.trim()}
          >
            <FiSend size={12} />
          </ActionIcon>
        </Group>
      )}
    </Stack>
  );
};

export default ReplySection;
