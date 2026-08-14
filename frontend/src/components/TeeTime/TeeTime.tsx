import dayjs from 'dayjs';
import { useNavigate } from 'react-router-dom';
import { Card, Text, Button, Group, Badge } from '@mantine/core';
import { FiCalendar, FiClock, FiUser, FiUsers, FiEdit2 } from 'react-icons/fi';
import { formatTeeTime } from '../../domain/teeTime/formatters';
import { isHost as checkIsHost } from '../../domain/teeTime/teeTimeService';
import type { Event, HandleInviteAction } from '../../types';

interface TeeTimeProps {
  type: string | undefined;
  event: Event;
  handleInviteAction: HandleInviteAction;
  currentUserId?: number;
}

const TeeTime = ({ type, event, handleInviteAction, currentUserId }: TeeTimeProps) => {
  const navigate = useNavigate();

  const filledSpots = event.open_spots - event.remaining_spots;
  const isHost = currentUserId ? checkIsHost(event, currentUserId) : false;

  return (
    <Card
      className='tee-time ff-card-hover'
      shadow='xs'
      withBorder
      style={{
        borderColor: 'var(--ff-border)',
        cursor: type === 'committed' && isHost ? 'pointer' : undefined,
      }}
      p='md'
      onClick={type === 'committed' && isHost ? () => navigate(`/edit-tee-time/${event.id}`) : undefined}
    >
      <Group justify='space-between' align='flex-start' mb='xs' wrap='nowrap'>
        <Text fw={700} style={{ color: 'var(--ff-heading)' }} size='md'>
          {event.course_name}
        </Text>
        {event.group_name && (
          <Badge size='sm' variant='light' color='forest'>
            {event.group_name}
          </Badge>
        )}
      </Group>

      <Group gap='lg' mb='xs'>
        <Group gap={6}>
          <FiCalendar size={14} style={{ color: 'var(--mantine-color-dimmed)' }} />
          <Text size='sm' c='dimmed'>{dayjs(event.date).format('MMM D')}</Text>
        </Group>
        <Group gap={6}>
          <FiClock size={14} style={{ color: 'var(--mantine-color-dimmed)' }} />
          <Text size='sm' c='dimmed'>{formatTeeTime(event.tee_time)}</Text>
        </Group>
      </Group>

      <Group gap='lg' mb='sm'>
        <Group gap={6}>
          <FiUser size={14} style={{ color: 'var(--mantine-color-dimmed)' }} />
          <Text size='sm' c='dimmed'>{event.host_name}</Text>
        </Group>
        <Group gap={6}>
          <FiUsers size={14} style={{ color: 'var(--mantine-color-dimmed)' }} />
          <Text size='sm' c='dimmed'>
            {filledSpots}/{event.open_spots} spots filled
          </Text>
        </Group>
      </Group>

      <Group justify='flex-end' gap='xs'>
        {type === 'committed' && isHost && (
          <Button
            className='edit'
            color='forest'
            variant='subtle'
            size='sm'
            leftSection={<FiEdit2 size={14} />}
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/edit-tee-time/${event.id}`);
            }}
          >
            Edit
          </Button>
        )}
        {type === 'committed' && (
          <Button
            className='cancel'
            color='red'
            variant='subtle'
            size='sm'
            onClick={(e) => {
              e.stopPropagation();
              handleInviteAction.cancel(event);
            }}
          >
            Cancel
          </Button>
        )}
        {type === 'available' && (
          <>
            <Button
              className='decline'
              color='red'
              variant='subtle'
              size='sm'
              onClick={() => handleInviteAction.update(event.id, 'declined')}
            >
              Decline
            </Button>
            <Button
              className='accept'
              color='forest'
              variant='filled'
              size='sm'
              onClick={() => handleInviteAction.update(event.id, 'accepted')}
            >
              Accept
            </Button>
          </>
        )}
        {type === 'joinable' && (
          <Button
            className='join'
            color='forest'
            variant='filled'
            size='sm'
            onClick={() => handleInviteAction.join(event.id)}
          >
            Join
          </Button>
        )}
      </Group>
    </Card>
  );
};

export default TeeTime;
