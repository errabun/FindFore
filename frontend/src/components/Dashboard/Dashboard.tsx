import { useCallback, useEffect, useState } from 'react';
import { Tabs, SimpleGrid, Paper, Text, Title, Group, Box } from '@mantine/core';
import { FiCalendar, FiMail, FiUsers } from 'react-icons/fi';
import PlayerList from '../PlayerList/PlayerList';
import TeeTimeContainer from '../TeeTimeContainer/TeeTimeContainer';
import Newsfeed from '../Newsfeed/Newsfeed';
import type { Event, Friend, Player, HandleFriends, HandleInviteAction } from '../../types';

interface DashboardProps {
  events: Event[];
  friendsEvents: Event[];
  currentUserId: number;
  currentUserName: string;
  screenWidth: number;
  handleInviteAction: HandleInviteAction;
  friends: Friend[];
  players: Player[];
  handleFriends: HandleFriends;
}

const Dashboard = ({
  events,
  friendsEvents,
  currentUserId,
  currentUserName,
  screenWidth,
  handleInviteAction,
  friends,
  players,
  handleFriends,
}: DashboardProps) => {
  const [availableTeeTimes, setAvailableTeeTimes] = useState<Event[]>([]);
  const [committedTeeTimes, setCommittedTeeTimes] = useState<Event[]>([]);
  const [activeTab, setActiveTab] = useState<string | null>('committed');

  const getAvailable = useCallback(() => {
    return events.filter((event) => {
      if (
        event.declined.includes(currentUserId) ||
        event.accepted.includes(currentUserId) ||
        event.closed.includes(currentUserId)
      ) {
        return false;
      } else if (
        event.pending.includes(currentUserId) ||
        !event.private
      ) {
        return true;
      }

      return false;
    });
  }, [events, currentUserId]);

  const getCommitted = useCallback(() => {
    return events.filter((event) =>
      event.accepted.includes(currentUserId)
    );
  }, [events, currentUserId]);

  useEffect(() => {
    setAvailableTeeTimes(getAvailable());
    setCommittedTeeTimes(getCommitted());
  }, [events, getAvailable, getCommitted]);

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good morning';
    if (hour < 17) return 'Good afternoon';
    return 'Good evening';
  };

  const firstName = currentUserName ? currentUserName.split(' ')[0] : '';

  return (
    <div className='dashboard' style={{ display: 'flex', flexDirection: screenWidth < 768 ? 'column' : 'row', width: '100%' }}>
      {screenWidth >= 1025 && (
        <PlayerList
          userId={currentUserId}
          screenWidth={screenWidth}
          friends={friends}
          players={players}
          handleFriends={handleFriends}
        />
      )}

      <Box style={{ flex: 1, overflow: 'auto' }} p={{ base: 'md', sm: 'xl' }}>
        {firstName && (
          <Box mb='lg'>
            <Title order={2} c='forest.9' fw={700}>
              {getGreeting()}, {firstName}
            </Title>
            <Text c='dimmed' size='sm' mt={4}>
              Here's what's happening with your tee times
            </Text>
          </Box>
        )}

        <SimpleGrid cols={{ base: 2, sm: 3 }} mb='xl' spacing='md'>
          <Paper p='md' shadow='xs'>
            <Group gap='xs' mb={4}>
              <FiCalendar style={{ color: '#2E5A2E' }} />
              <Text size='xs' c='dimmed' fw={500}>Upcoming Rounds</Text>
            </Group>
            <Text size='xl' fw={700} c='forest.6'>
              {committedTeeTimes.length}
            </Text>
          </Paper>
          <Paper p='md' shadow='xs'>
            <Group gap='xs' mb={4}>
              <FiMail style={{ color: '#2E5A2E' }} />
              <Text size='xs' c='dimmed' fw={500}>Available Invites</Text>
            </Group>
            <Text size='xl' fw={700} c='forest.6'>
              {availableTeeTimes.length}
            </Text>
          </Paper>
          <Paper p='md' shadow='xs'>
            <Group gap='xs' mb={4}>
              <FiUsers style={{ color: '#2E5A2E' }} />
              <Text size='xs' c='dimmed' fw={500}>Friends</Text>
            </Group>
            <Text size='xl' fw={700} c='forest.6'>
              {friends.length}
            </Text>
          </Paper>
        </SimpleGrid>

        {screenWidth < 768 && (
          <Tabs value={activeTab} onChange={setActiveTab} mb='md' color='forest'>
            <Tabs.List grow>
              <Tabs.Tab value='committed'>Committed</Tabs.Tab>
              <Tabs.Tab value='available'>Available</Tabs.Tab>
              <Tabs.Tab value='feed'>Feed</Tabs.Tab>
            </Tabs.List>
          </Tabs>
        )}

        {screenWidth >= 768 && (
          <>
            <SimpleGrid cols={2} spacing='md' mb='md'>
              <TeeTimeContainer
                title='Committed Tee Times'
                events={committedTeeTimes}
                handleInviteAction={handleInviteAction}
                currentUserId={currentUserId}
              />
              <TeeTimeContainer
                title='Available Tee Times'
                events={availableTeeTimes}
                friendsEvents={friendsEvents}
                friendIds={friends.map((f) => f.id)}
                handleInviteAction={handleInviteAction}
              />
            </SimpleGrid>
            <Newsfeed currentUserId={currentUserId} currentUserName={currentUserName} />
          </>
        )}

        {activeTab === 'committed' && screenWidth < 768 && (
          <TeeTimeContainer
            title='Committed Tee Times'
            events={committedTeeTimes}
            handleInviteAction={handleInviteAction}
            currentUserId={currentUserId}
          />
        )}
        {activeTab === 'available' && screenWidth < 768 && (
          <TeeTimeContainer
            title='Available Tee Times'
            events={availableTeeTimes}
            friendsEvents={friendsEvents}
            friendIds={friends.map((f) => f.id)}
            handleInviteAction={handleInviteAction}
          />
        )}
        {activeTab === 'feed' && screenWidth < 768 && (
          <Newsfeed currentUserId={currentUserId} currentUserName={currentUserName} />
        )}
      </Box>
    </div>
  );
};

export default Dashboard;
