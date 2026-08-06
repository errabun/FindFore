import { Paper, Group, Text, ActionIcon, Avatar, Button, Stack } from '@mantine/core';
import { FiUserPlus, FiUserMinus, FiCheck, FiX, FiClock } from 'react-icons/fi';
import type { Friend, FriendRequest, Player, HandleFriends } from '../../types';

interface PlayerCardProps {
  playerInfo: Friend | Player;
  friends: Friend[];
  incomingRequests: FriendRequest[];
  outgoingPendingIds: number[];
  handleFriends: HandleFriends;
}

const getInitials = (name: string) => {
  const parts = name.split(' ');
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  }
  return name.substring(0, 2).toUpperCase();
};

type Relationship =
  | { kind: 'friend'; friend: Friend }
  | { kind: 'incoming'; request: FriendRequest }
  | { kind: 'outgoing' }
  | { kind: 'none' };

function resolveRelationship(
  playerId: number,
  friends: Friend[],
  incomingRequests: FriendRequest[],
  outgoingPendingIds: number[],
): Relationship {
  const friend = friends.find((f) => f.id === playerId);
  if (friend) return { kind: 'friend', friend };

  const incoming = incomingRequests.find((r) => r.requesterId === playerId);
  if (incoming) return { kind: 'incoming', request: incoming };

  if (outgoingPendingIds.includes(playerId)) return { kind: 'outgoing' };

  return { kind: 'none' };
}

const PlayerCard = ({
  playerInfo,
  friends,
  incomingRequests,
  outgoingPendingIds,
  handleFriends,
}: PlayerCardProps) => {
  const relationship = resolveRelationship(
    playerInfo.id,
    friends,
    incomingRequests,
    outgoingPendingIds,
  );

  return (
    <Paper
      data-cy='player-card'
      className='player-card'
      shadow='xs'
      radius='md'
      p='sm'
      style={{
        transition: 'background-color 0.15s ease',
        cursor: 'default',
      }}
      onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'var(--mantine-color-sand-1)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = ''; }}
    >
      <Group justify='space-between' wrap='nowrap'>
        <Group gap='sm' wrap='nowrap'>
          <Avatar size='sm' radius='xl' color='forest' variant='filled'>
            {getInitials(playerInfo.name)}
          </Avatar>
          <Text size='sm' fw={500} truncate>
            {playerInfo.name}
          </Text>
        </Group>

        {relationship.kind === 'friend' && (
          <ActionIcon
            data-cy='friend-option'
            variant='subtle'
            color='red'
            size='sm'
            radius='xl'
            onClick={() => handleFriends.remove(relationship.friend)}
            title='Remove Friend'
          >
            <FiUserMinus size={14} />
          </ActionIcon>
        )}

        {relationship.kind === 'outgoing' && (
          <ActionIcon
            variant='light'
            color='gray'
            size='sm'
            radius='xl'
            disabled
            title='Request pending'
          >
            <FiClock size={14} />
          </ActionIcon>
        )}

        {relationship.kind === 'incoming' && (
          <Group gap={4} wrap='nowrap'>
            <ActionIcon
              variant='filled'
              color='forest'
              size='sm'
              radius='xl'
              onClick={() => handleFriends.accept(relationship.request.id)}
              title='Accept request'
            >
              <FiCheck size={14} />
            </ActionIcon>
            <ActionIcon
              variant='subtle'
              color='red'
              size='sm'
              radius='xl'
              onClick={() => handleFriends.decline(relationship.request.id)}
              title='Decline request'
            >
              <FiX size={14} />
            </ActionIcon>
          </Group>
        )}

        {relationship.kind === 'none' && (
          <ActionIcon
            data-cy='friend-option'
            variant='filled'
            color='forest'
            size='sm'
            radius='xl'
            onClick={() => handleFriends.request(playerInfo)}
            title='Request Friend'
          >
            <FiUserPlus size={14} />
          </ActionIcon>
        )}
      </Group>
    </Paper>
  );
};

interface FriendRequestCardProps {
  request: FriendRequest;
  onAccept: (requestId: number) => void;
  onDecline: (requestId: number) => void;
  compact?: boolean;
}

export const FriendRequestCard = ({
  request,
  onAccept,
  onDecline,
  compact = false,
}: FriendRequestCardProps) => {
  const actions = (
    <Group gap='xs' wrap='nowrap' grow={compact}>
      <Button
        size='xs'
        color='forest'
        onClick={() => onAccept(request.id)}
      >
        Accept
      </Button>
      <Button
        size='xs'
        variant='subtle'
        color='gray'
        onClick={() => onDecline(request.id)}
      >
        Decline
      </Button>
    </Group>
  );

  // Sidebar Requests tab is narrow — stack actions under the name so it isn't truncated.
  if (compact) {
    return (
      <Paper shadow='xs' p='sm' radius='md'>
        <Stack gap='sm'>
          <Group gap='sm' wrap='nowrap' align='flex-start'>
            <Avatar size='sm' radius='xl' color='forest' variant='filled'>
              {getInitials(request.requesterName)}
            </Avatar>
            <Stack gap={0} style={{ flex: 1, minWidth: 0 }}>
              <Text size='sm' fw={600} style={{ wordBreak: 'break-word' }}>
                {request.requesterName}
              </Text>
              <Text size='xs' c='dimmed'>
                wants to be friends
              </Text>
            </Stack>
          </Group>
          {actions}
        </Stack>
      </Paper>
    );
  }

  return (
    <Paper shadow='sm' p='md' radius='md'>
      <Group justify='space-between' wrap='nowrap' align='center'>
        <Group gap='sm' wrap='nowrap' style={{ minWidth: 0, flex: 1 }}>
          <Avatar size='md' radius='xl' color='forest' variant='filled'>
            {getInitials(request.requesterName)}
          </Avatar>
          <Stack gap={0} style={{ minWidth: 0 }}>
            <Text size='sm' fw={600} style={{ wordBreak: 'break-word' }}>
              {request.requesterName}
            </Text>
            <Text size='xs' c='dimmed'>
              wants to be friends
            </Text>
          </Stack>
        </Group>
        {actions}
      </Group>
    </Paper>
  );
};

export default PlayerCard;
