import { useState } from 'react';
import { SegmentedControl, Text, Stack, Button, ThemeIcon } from '@mantine/core';
import { FiUsers, FiMail } from 'react-icons/fi';
import PlayerCard, { FriendRequestCard } from '../PlayerCard/PlayerCard';
import type { Friend, FriendRequest, Player, HandleFriends } from '../../types';

interface PlayerListProps {
  screenWidth: number;
  players: Player[];
  friends: Friend[];
  incomingRequests: FriendRequest[];
  outgoingPendingIds: number[];
  handleFriends: HandleFriends;
  userId: number;
}

const PlayerList = ({
  screenWidth,
  players,
  friends,
  incomingRequests,
  outgoingPendingIds,
  handleFriends,
  userId,
}: PlayerListProps) => {
  const [playerType, setPlayerType] = useState('friends');

  const mapPlayers = (type: (Friend | Player)[]) => {
    return type
      .filter((t) => t.id !== userId)
      .map((p) => (
        <PlayerCard
          key={p.id}
          playerInfo={p}
          friends={friends}
          incomingRequests={incomingRequests}
          outgoingPendingIds={outgoingPendingIds}
          handleFriends={handleFriends}
        />
      ));
  };

  const isDesktop = screenWidth > 1023;

  return (
    <aside
      data-cy='player-list'
      style={{
        padding: isDesktop ? '1.5rem' : '2rem',
        height: isDesktop ? '100%' : '100vh',
        width: isDesktop ? 320 : '100%',
        backgroundColor: 'var(--ff-bg)',
        display: 'flex',
        flexDirection: 'column',
        borderRight: isDesktop ? '1px solid var(--ff-border)' : 'none',
        overflowY: 'auto',
      }}
    >
      <SegmentedControl
        value={playerType}
        onChange={setPlayerType}
        mb='lg'
        color='forest'
        fullWidth
        data={[
          { label: 'Friends', value: 'friends' },
          {
            label: incomingRequests.length
              ? `Requests (${incomingRequests.length})`
              : 'Requests',
            value: 'requests',
          },
          { label: 'Community', value: 'community' },
        ]}
        data-cy='player-type'
      />
      <Stack gap='xs'>
        {playerType === 'friends' && !friends.length && (
          <Stack align='center' gap='md' py='xl'>
            <ThemeIcon size='xl' radius='xl' variant='light' color='forest'>
              <FiUsers size={20} />
            </ThemeIcon>
            <Text ta='center' c='dimmed' size='sm'>
              You don't have any friends yet.
              <br />
              Browse the community to connect!
            </Text>
            <Button
              variant='light'
              color='forest'
              size='sm'
              onClick={() => setPlayerType('community')}
            >
              Browse Community
            </Button>
          </Stack>
        )}

        {playerType === 'friends' && mapPlayers(friends)}

        {playerType === 'requests' && !incomingRequests.length && (
          <Stack align='center' gap='md' py='xl'>
            <ThemeIcon size='xl' radius='xl' variant='light' color='forest'>
              <FiMail size={20} />
            </ThemeIcon>
            <Text ta='center' c='dimmed' size='sm'>
              No pending friend requests.
            </Text>
          </Stack>
        )}

        {playerType === 'requests' &&
          incomingRequests.map((request) => (
            <FriendRequestCard
              key={request.id}
              request={request}
              onAccept={handleFriends.accept}
              onDecline={handleFriends.decline}
              compact
            />
          ))}

        {playerType === 'community' && mapPlayers(players)}
      </Stack>
    </aside>
  );
};

export default PlayerList;
