import { Card, Group, Text, Avatar, Menu, ActionIcon, Tooltip, Popover } from '@mantine/core';
import { useState } from 'react';
import { FiMoreHorizontal, FiTrash2, FiMessageCircle, FiPlus } from 'react-icons/fi';
import dayjs from 'dayjs';
import ReplySection from './ReplySection';
import type { Post, Reaction } from '../../types';
import classes from './PostCard.module.css';

const EMOJI_OPTIONS = [
  { emoji: '\u{1F44D}', label: 'Thumbs Up' },
  { emoji: '\u{1F525}', label: 'Fire' },
  { emoji: '\u{1F44F}', label: 'Clap' },
  { emoji: '\u2764\uFE0F', label: 'Heart' },
  { emoji: '\u26F3', label: 'Golf' },
  { emoji: '\u{1F602}', label: 'Laugh' },
];

interface GroupedReaction {
  emoji: string;
  count: number;
  playerNames: string[];
  reacted: boolean;
}

function groupReactions(reactions: Reaction[], currentUserId: number): GroupedReaction[] {
  const grouped: GroupedReaction[] = [];
  const emojiMap = new Map<string, GroupedReaction>();

  for (const r of reactions) {
    const existing = emojiMap.get(r.emoji);
    if (existing) {
      existing.count++;
      existing.playerNames.push(r.player_name);
      if (r.player_id === currentUserId) existing.reacted = true;
    } else {
      const entry: GroupedReaction = {
        emoji: r.emoji,
        count: 1,
        playerNames: [r.player_name],
        reacted: r.player_id === currentUserId,
      };
      emojiMap.set(r.emoji, entry);
      grouped.push(entry);
    }
  }

  return grouped;
}

interface PostCardProps {
  post: Post;
  currentUserId: number;
  canDelete?: boolean;
  onToggleReaction: (postId: number, emoji: string) => void;
  onCreateReply: (postId: number, body: string) => void;
  onDeletePost: (postId: number) => void;
  onDeleteReply: (postId: number, replyId: number) => void;
}

const PostCard = ({
  post,
  currentUserId,
  canDelete,
  onToggleReaction,
  onCreateReply,
  onDeletePost,
  onDeleteReply,
}: PostCardProps) => {
  const [pickerOpen, setPickerOpen] = useState(false);
  const [showReplies, setShowReplies] = useState(false);
  const isAuthor = post.player_id === currentUserId;
  const showDelete = canDelete ?? isAuthor;
  const grouped = groupReactions(post.reactions, currentUserId);
  const reactionCount = post.reactions.length;
  const replyCount = post.replies.length;

  const reactionText = reactionCount > 0
    ? `${reactionCount} ${reactionCount === 1 ? 'reaction' : 'reactions'}`
    : null;
  const replyText = replyCount > 0
    ? `${replyCount} ${replyCount === 1 ? 'reply' : 'replies'}`
    : null;
  const reactionTooltip = grouped.map((g) => `${g.emoji} ${g.count}`).join('  ');

  return (
    <Card withBorder padding='lg' radius='md' className={classes.card}>
      <Group justify='space-between'>
        <Group gap='sm'>
          <Avatar size='sm' radius='sm' color='forest'>
            {post.player_name.charAt(0)}
          </Avatar>
          <div>
            <Text c='bright' fw={500} size='sm'>
              {post.player_name}
            </Text>
            <Text size='xs' c='dimmed'>
              {dayjs(post.created_at).format('MMM D [at] h:mm A')}
            </Text>
          </div>
        </Group>
        {showDelete && (
          <Menu position='bottom-end' withArrow>
            <Menu.Target>
              <ActionIcon variant='subtle' color='gray' size='sm'>
                <FiMoreHorizontal size={16} />
              </ActionIcon>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Item
                color='red'
                leftSection={<FiTrash2 size={14} />}
                onClick={() => onDeletePost(post.id)}
              >
                Delete post
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        )}
      </Group>

      <Text className={classes.body} size='sm'>
        {post.body}
      </Text>

      <Card.Section className={classes.footer}>
        <Group justify='space-between'>
          <Text size='xs' c='dimmed'>
            {reactionText ? (
              <Tooltip label={reactionTooltip} withArrow position='bottom'>
                <span style={{ cursor: 'default' }}>{reactionText}</span>
              </Tooltip>
            ) : null}
            {reactionText && replyText ? ' \u00B7 ' : null}
            {replyText ?? (!reactionText ? 'Be the first to react' : null)}
          </Text>
          <Group gap={0}>
            {grouped.map((g) => (
              <Tooltip key={g.emoji} label={g.playerNames.join(', ')} withArrow>
                <ActionIcon
                  variant='subtle'
                  color={g.reacted ? 'forest' : 'gray'}
                  onClick={() => onToggleReaction(post.id, g.emoji)}
                  style={{ fontSize: 16 }}
                >
                  <span>{g.emoji}</span>
                  {g.count > 1 && (
                    <Text size='xs' ml={2} fw={500} c={g.reacted ? 'forest.6' : 'dimmed'}>
                      {g.count}
                    </Text>
                  )}
                </ActionIcon>
              </Tooltip>
            ))}
            <Popover opened={pickerOpen} onChange={setPickerOpen} position='bottom-end' withArrow>
              <Popover.Target>
                <ActionIcon
                  variant='subtle'
                  color='gray'
                  onClick={() => setPickerOpen(!pickerOpen)}
                >
                  <FiPlus size={16} />
                </ActionIcon>
              </Popover.Target>
              <Popover.Dropdown p='xs'>
                <Group gap={4}>
                  {EMOJI_OPTIONS.map((opt) => (
                    <Tooltip key={opt.emoji} label={opt.label} withArrow>
                      <ActionIcon
                        variant='subtle'
                        size='lg'
                        onClick={() => {
                          onToggleReaction(post.id, opt.emoji);
                          setPickerOpen(false);
                        }}
                        style={{ fontSize: 20 }}
                      >
                        {opt.emoji}
                      </ActionIcon>
                    </Tooltip>
                  ))}
                </Group>
              </Popover.Dropdown>
            </Popover>
            <ActionIcon
              variant='subtle'
              color={showReplies ? 'forest' : 'gray'}
              onClick={() => setShowReplies(!showReplies)}
            >
              <FiMessageCircle size={16} />
            </ActionIcon>
          </Group>
        </Group>
      </Card.Section>

      {showReplies && (
        <Card.Section p='md' pt='xs'>
          <ReplySection
            replies={post.replies}
            currentUserId={currentUserId}
            onCreateReply={(body) => onCreateReply(post.id, body)}
            onDeleteReply={(replyId) => onDeleteReply(post.id, replyId)}
            defaultShowInput
          />
        </Card.Section>
      )}
    </Card>
  );
};

export default PostCard;
