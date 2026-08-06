import { useState } from 'react';
import { Tabs, SimpleGrid, Paper, Text, Title, Group, Box, Stack } from '@mantine/core';
import { FiCalendar, FiMail, FiUsers } from 'react-icons/fi';
import PlayerList from '../PlayerList/PlayerList';
import TeeTimeContainer from '../TeeTimeContainer/TeeTimeContainer';
import Newsfeed from '../Newsfeed/Newsfeed';
import { FriendRequestCard } from '../PlayerCard/PlayerCard';
import { filterCommitted, filterAvailable } from '../../domain/teeTime/teeTimeService';
import type {
  Event,
  Friend,
  FriendRequest,
  Player,
  HandleFriends,
  HandleInviteAction,
} from '../../types';

interface DashboardProps {
  events: Event[];
  friendsEvents: Event[];
  currentUserId: number;
  currentUserName: string;
  screenWidth: number;
  handleInviteAction: HandleInviteAction;
  friends: Friend[];
  players: Player[];
  incomingRequests: FriendRequest[];
  outgoingPendingIds: number[];
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
  incomingRequests,
  outgoingPendingIds,
  handleFriends,
}: DashboardProps) => {
  const [activeTab, setActiveTab] = useState<string | null>('committed');

  const committedTeeTimes = filterCommitted(events, currentUserId);
  const availableTeeTimes = filterAvailable(events, currentUserId);
  const friendIds = friends.map((f) => f.id);

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
          incomingRequests={incomingRequests}
          outgoingPendingIds={outgoingPendingIds}
          handleFriends={handleFriends}
        />
      )}

      <Box style={{ flex: 1, overflow: 'auto' }} p={{ base: 'md', sm: 'xl' }}>
        {firstName && (
          <Box mb='lg'>
            <Title order={2} style={{ color: 'var(--ff-heading)' }} fw={700}>
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
              <FiCalendar style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed' fw={500}>Upcoming Rounds</Text>
            </Group>
            <Text size='xl' fw={700} style={{ color: 'var(--ff-stat)' }}>
              {committedTeeTimes.length}
            </Text>
          </Paper>
          <Paper p='md' shadow='xs'>
            <Group gap='xs' mb={4}>
              <FiMail style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed' fw={500}>Available Invites</Text>
            </Group>
            <Text size='xl' fw={700} style={{ color: 'var(--ff-stat)' }}>
              {availableTeeTimes.length}
            </Text>
          </Paper>
          <Paper p='md' shadow='xs'>
            <Group gap='xs' mb={4}>
              <FiUsers style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed' fw={500}>Friends</Text>
            </Group>
            <Text size='xl' fw={700} style={{ color: 'var(--ff-stat)' }}>
              {friends.length}
            </Text>
          </Paper>
        </SimpleGrid>

        {incomingRequests.length > 0 && (
          <Paper p='md' shadow='sm' mb='md' style={{ border: '1px solid var(--ff-border)' }}>
            <Group gap='xs' mb='sm'>
              <FiUsers style={{ color: 'var(--ff-icon-primary)' }} />
              <Title order={5} style={{ color: 'var(--ff-heading)' }}>
                Friend requests
              </Title>
              <Text size='sm' c='dimmed'>
                ({incomingRequests.length})
              </Text>
            </Group>
            <Stack gap='sm'>
              {incomingRequests.map((request) => (
                <FriendRequestCard
                  key={request.id}
                  request={request}
                  onAccept={handleFriends.accept}
                  onDecline={handleFriends.decline}
                />
              ))}
            </Stack>
          </Paper>
        )}

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
                friendIds={friendIds}
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
            friendIds={friendIds}
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
