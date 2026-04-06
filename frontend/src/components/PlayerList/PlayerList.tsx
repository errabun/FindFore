import { useState } from 'react';
import { SegmentedControl, Text, Stack, Button, ThemeIcon } from '@mantine/core';
import { FiUsers } from 'react-icons/fi';
import PlayerCard from '../PlayerCard/PlayerCard';
import type { Friend, Player, HandleFriends } from '../../types';

interface PlayerListProps {
  screenWidth: number;
  players: Player[];
  friends: Friend[];
  handleFriends: HandleFriends;
  userId: number;
}

const PlayerList = ({ screenWidth, players, friends, handleFriends, userId }: PlayerListProps) => {
  const [playerType, setPlayerType] = useState('friends');

  const mapPlayers = (type: (Friend | Player)[]) => {
    return type
      .filter((t) => t.id !== userId)
      .map((p) => (
        <PlayerCard
          key={p.id}
          playerInfo={p}
          friends={friends}
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
          { label: 'Community', value: 'community' },
        ]}
        data-cy='player-type'
      />
      <Stack gap='xs'>
        {!friends.length && playerType === 'friends' && (
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
        {playerType === 'friends' ? mapPlayers(friends) : mapPlayers(players)}
      </Stack>
    </aside>
  );
};

export default PlayerList;
