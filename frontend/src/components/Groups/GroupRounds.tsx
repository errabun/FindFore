import { Link } from 'react-router-dom';
import { Alert, Button, Group, Skeleton, Stack } from '@mantine/core';
import { FiAlertCircle, FiCalendar, FiPlus } from 'react-icons/fi';
import EmptyState from '../EmptyState/EmptyState';
import TeeTime from '../TeeTime/TeeTime';
import { useGroupRounds } from '../../hooks/useGroupRounds';

interface GroupRoundsProps {
  groupId: number;
  groupName: string;
  currentUserId: number;
}

function roundType(event: { accepted: number[]; remaining_spots: number }, currentUserId: number) {
  if (event.accepted?.includes(currentUserId)) return 'committed';
  if (event.remaining_spots > 0) return 'joinable';
  return undefined;
}

export default function GroupRounds({ groupId, groupName, currentUserId }: GroupRoundsProps) {
  const { events, loading, error, join, cancel } = useGroupRounds(groupId, currentUserId);
  const planHref = `/event-form?group=${groupId}&name=${encodeURIComponent(groupName)}`;

  return (
    <Stack gap='md'>
      {error && (
        <Alert color='red' icon={<FiAlertCircle />}>
          {error}
        </Alert>
      )}
      {events.length > 0 && (
        <Group justify='flex-end'>
          <Button
            component={Link}
            to={planHref}
            color='forest'
            size='md'
            leftSection={<FiPlus size={14} />}
          >
            Plan a round
          </Button>
        </Group>
      )}

      {loading && events.length === 0 ? (
        <Stack gap='sm'>
          <Skeleton height={120} radius='md' />
          <Skeleton height={80} radius='md' />
        </Stack>
      ) : events.length === 0 ? (
        <EmptyState
          icon={<FiCalendar />}
          title='No upcoming rounds'
          description='Plan one for the group. Members can join from this page.'
          actionLabel='Plan a round'
          actionHref={planHref}
        />
      ) : (
        events.map((event) => (
          <TeeTime
            key={event.id}
            type={roundType(event, currentUserId)}
            event={event}
            currentUserId={currentUserId}
            handleInviteAction={{
              update: () => {},
              cancel,
              join,
            }}
          />
        ))
      )}
    </Stack>
  );
}
