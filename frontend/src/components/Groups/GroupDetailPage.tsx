import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  Badge,
  Button,
  Container,
  Group,
  Select,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import { FiArrowLeft } from 'react-icons/fi';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { useGroup } from '../../hooks/useGroups';
import type { Player } from '../../domain/auth/types';

interface GroupDetailPageProps {
  hostPlayer: number;
  players: Player[];
}

export default function GroupDetailPage({ hostPlayer, players }: GroupDetailPageProps) {
  const { groupId } = useParams();
  const id = Number(groupId);
  const navigate = useNavigate();
  const { group, members, joinRequests, refresh } = useGroup(id, hostPlayer);
  const [invitee, setInvitee] = useState<string | null>(null);

  if (!group) {
    return (
      <Container size='sm' py='lg'>
        <Text c='dimmed'>Group not found.</Text>
      </Container>
    );
  }

  const viewer = group.viewer_membership;
  const isActive = viewer?.status === 'active';
  const isPending = viewer?.status === 'pending';
  const canManage = isActive && (viewer?.role === 'owner' || viewer?.role === 'admin');
  const isOwner = viewer?.role === 'owner';

  const join = () => groupAdapter.join(group.id).then(() => refresh());
  const leave = () => groupAdapter.leave(group.id).then(() => navigate('/groups'));
  const invite = () => {
    if (!invitee) return;
    groupAdapter.invite(group.id, Number(invitee)).then(() => {
      setInvitee(null);
      refresh();
    });
  };

  const inviteOptions = players
    .filter((p) => p.id !== hostPlayer && !members.some((m) => m.player_id === p.id))
    .map((p) => ({ value: String(p.id), label: p.name }));

  return (
    <Container size='sm' py='lg'>
      <Button
        component={Link}
        to='/groups'
        variant='subtle'
        color='gray'
        leftSection={<FiArrowLeft />}
        mb='md'
        size='sm'
      >
        Groups
      </Button>

      <Group justify='space-between' align='flex-start' mb='sm'>
        <div>
          <Title order={2}>{group.name}</Title>
          <Text size='sm' c='dimmed'>
            {group.member_count} members · {group.privacy === 'public' ? 'Public' : 'Private'}
          </Text>
        </div>
        {isActive && <Badge color='teal'>Joined</Badge>}
        {isPending && <Badge color='yellow'>Request pending</Badge>}
      </Group>

      {group.description && (
        <Stack gap={4} mb='md'>
          <Text fw={600} size='sm'>
            About
          </Text>
          <Text size='sm'>{group.description}</Text>
        </Stack>
      )}

      {!isActive && !isPending && (
        <Button color='forest' mb='md' onClick={join}>
          {group.privacy === 'public' ? 'Join Group' : 'Request to Join'}
        </Button>
      )}
      {isPending && (
        <Button variant='light' color='gray' mb='md' onClick={leave}>
          Cancel request
        </Button>
      )}
      {isActive && !isOwner && (
        <Button variant='light' color='red' mb='md' onClick={leave}>
          Leave group
        </Button>
      )}

      {isActive && (
        <Stack gap='xs' mb='lg'>
          <Text fw={600} size='sm'>
            Members
          </Text>
          {members.map((m) => (
            <Group key={m.player_id} justify='space-between'>
              <Text size='sm'>
                {m.player_name || `Player ${m.player_id}`}
                {m.role !== 'member' ? ` · ${m.role}` : ''}
              </Text>
              {canManage && m.role === 'member' && (
                <Button
                  size='compact-xs'
                  variant='subtle'
                  color='red'
                  onClick={() => groupAdapter.removeMember(group.id, m.player_id).then(() => refresh())}
                >
                  Remove
                </Button>
              )}
            </Group>
          ))}
        </Stack>
      )}

      {canManage && joinRequests.length > 0 && (
        <Stack gap='xs' mb='lg'>
          <Text fw={600} size='sm'>
            Join requests
          </Text>
          {joinRequests.map((req) => (
            <Group key={req.player_id} justify='space-between'>
              <Text size='sm'>{req.player_name || `Player ${req.player_id}`}</Text>
              <Group gap='xs'>
                <Button
                  size='compact-xs'
                  color='forest'
                  onClick={() => groupAdapter.approveJoinRequest(group.id, req.player_id).then(() => refresh())}
                >
                  Approve
                </Button>
                <Button
                  size='compact-xs'
                  variant='light'
                  onClick={() => groupAdapter.denyJoinRequest(group.id, req.player_id).then(() => refresh())}
                >
                  Deny
                </Button>
              </Group>
            </Group>
          ))}
        </Stack>
      )}

      {canManage && (
        <Stack gap='xs' mb='lg'>
          <Text fw={600} size='sm'>
            Invite a golfer
          </Text>
          <Group align='flex-end'>
            <Select
              placeholder='Choose a player'
              data={inviteOptions}
              value={invitee}
              onChange={setInvitee}
              searchable
              style={{ flex: 1 }}
            />
            <Button color='forest' size='sm' disabled={!invitee} onClick={invite}>
              Invite
            </Button>
          </Group>
        </Stack>
      )}

      <Text size='sm' c='dimmed' mt='xl'>
        Upcoming — Coming soon
      </Text>
    </Container>
  );
}
