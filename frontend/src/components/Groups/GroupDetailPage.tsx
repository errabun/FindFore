import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  Alert,
  Badge,
  Button,
  Container,
  Group,
  Modal,
  Select,
  Skeleton,
  Stack,
  Tabs,
  Text,
  Title,
} from '@mantine/core';
import { FiAlertCircle, FiArrowLeft, FiUsers } from 'react-icons/fi';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';
import { useGroup } from '../../hooks/useGroups';
import type { Friend } from '../../domain/social/types';
import GroupSettingsPanel from './GroupSettingsPanel';

interface GroupDetailPageProps {
  hostPlayer: number;
  friends: Friend[];
}

function actionMessage(err: unknown, fallback: string) {
  return err instanceof ApiError ? err.message : fallback;
}

export default function GroupDetailPage({ hostPlayer, friends }: GroupDetailPageProps) {
  const { groupId } = useParams();
  const id = Number(groupId);
  const navigate = useNavigate();
  const { group, members, joinRequests, invitations, loading, error, notFound, refresh } = useGroup(
    id,
    hostPlayer,
  );
  const [invitee, setInvitee] = useState<string | null>(null);
  const [actionError, setActionError] = useState('');
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState<
    | { kind: 'leave' }
    | { kind: 'remove'; playerId: number; name: string }
    | { kind: 'cancelInvite'; invitationId: number; name: string }
    | null
  >(null);

  if (loading && !group) {
    return (
      <Container size='sm' py='lg'>
        <Skeleton height={28} width={120} mb='md' />
        <Skeleton height={36} mb='sm' />
        <Skeleton height={80} />
      </Container>
    );
  }

  if (notFound || !group) {
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
        <Alert color={error ? 'red' : 'gray'} icon={<FiAlertCircle />}>
          {error || 'Group not found.'}
        </Alert>
      </Container>
    );
  }

  const viewer = group.viewer_membership;
  const isActive = viewer?.status === 'active';
  const isPending = viewer?.status === 'pending';
  const canManage = isActive && (viewer?.role === 'owner' || viewer?.role === 'admin');
  const isOwner = viewer?.role === 'owner';

  const run = async (fn: () => Promise<unknown>, fallback: string) => {
    setBusy(true);
    setActionError('');
    try {
      await fn();
      await refresh();
    } catch (err) {
      setActionError(actionMessage(err, fallback));
    } finally {
      setBusy(false);
    }
  };

  const memberIds = new Set(members.map((m) => m.player_id));
  const invitedIds = new Set(invitations.map((i) => i.invitee_player_id));
  const inviteOptions = friends
    .filter((f) => f.id !== hostPlayer && !memberIds.has(f.id) && !invitedIds.has(f.id))
    .map((f) => ({ value: String(f.id), label: f.name }));

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

      {(error || actionError) && (
        <Alert color='red' icon={<FiAlertCircle />} mb='md'>
          {actionError || error}
        </Alert>
      )}

      <Group justify='space-between' align='flex-start' mb='sm'>
        <div>
          <Title order={2}>{group.name}</Title>
          <Text size='sm' c='dimmed'>
            {group.member_count} {group.member_count === 1 ? 'member' : 'members'} ·{' '}
            {group.privacy === 'public' ? 'Public' : 'Private'}
            {group.owner?.name ? ` · ${group.owner.name}` : ''}
          </Text>
        </div>
        {isActive && <Badge color='teal'>Joined</Badge>}
        {isPending && <Badge color='yellow'>Request pending</Badge>}
      </Group>

      {group.description && (
        <Text size='sm' mb='md'>
          {group.description}
        </Text>
      )}

      {!isActive && !isPending && (
        <Button
          color='forest'
          mb='md'
          size='md'
          loading={busy}
          onClick={() => run(() => groupAdapter.join(group.id), 'Could not join group')}
        >
          {group.privacy === 'public' ? 'Join Group' : 'Request to Join'}
        </Button>
      )}
      {isPending && (
        <Button
          variant='light'
          color='gray'
          mb='md'
          size='md'
          loading={busy}
          onClick={() =>
            run(() => groupAdapter.leave(group.id).then(() => navigate('/groups')), 'Could not cancel request')
          }
        >
          Cancel request
        </Button>
      )}
      {isActive && !isOwner && (
        <Button variant='light' color='red' mb='md' size='md' onClick={() => setConfirm({ kind: 'leave' })}>
          Leave group
        </Button>
      )}

      {isActive && (
        <Tabs defaultValue='members' mt='md'>
          <Tabs.List>
            <Tabs.Tab value='members' leftSection={<FiUsers size={14} />}>
              Members
            </Tabs.Tab>
            {canManage && <Tabs.Tab value='settings'>Settings</Tabs.Tab>}
          </Tabs.List>

          <Tabs.Panel value='members' pt='md'>
            <Stack gap='xs' mb='lg'>
              {members.map((m) => (
                <Group key={m.player_id} justify='space-between' wrap='nowrap'>
                  <Text size='sm'>
                    {m.player_name || `Player ${m.player_id}`}
                    {m.role !== 'member' ? ` · ${m.role}` : ''}
                  </Text>
                  {canManage && m.role === 'member' && (
                    <Button
                      size='sm'
                      variant='subtle'
                      color='red'
                      onClick={() =>
                        setConfirm({
                          kind: 'remove',
                          playerId: m.player_id,
                          name: m.player_name || `Player ${m.player_id}`,
                        })
                      }
                    >
                      Remove
                    </Button>
                  )}
                </Group>
              ))}
            </Stack>

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
                        size='sm'
                        color='forest'
                        loading={busy}
                        onClick={() =>
                          run(
                            () => groupAdapter.approveJoinRequest(group.id, req.player_id),
                            'Could not approve request',
                          )
                        }
                      >
                        Approve
                      </Button>
                      <Button
                        size='sm'
                        variant='light'
                        loading={busy}
                        onClick={() =>
                          run(
                            () => groupAdapter.denyJoinRequest(group.id, req.player_id),
                            'Could not deny request',
                          )
                        }
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
                  Invite a friend
                </Text>
                {invitations.length > 0 && (
                  <Stack gap='xs'>
                    {invitations.map((inv) => (
                      <Group key={inv.id} justify='space-between'>
                        <Text size='sm'>{inv.invitee_name || `Player ${inv.invitee_player_id}`}</Text>
                        <Button
                          size='sm'
                          variant='subtle'
                          color='gray'
                          onClick={() =>
                            setConfirm({
                              kind: 'cancelInvite',
                              invitationId: inv.id,
                              name: inv.invitee_name || `Player ${inv.invitee_player_id}`,
                            })
                          }
                        >
                          Cancel
                        </Button>
                      </Group>
                    ))}
                  </Stack>
                )}
                {inviteOptions.length === 0 ? (
                  <Text size='sm' c='dimmed'>
                    No friends left to invite. Add friends first, then invite them here.
                  </Text>
                ) : (
                  <Group align='flex-end'>
                    <Select
                      label='Friend'
                      placeholder='Choose a friend'
                      data={inviteOptions}
                      value={invitee}
                      onChange={setInvitee}
                      searchable
                      style={{ flex: 1 }}
                    />
                    <Button
                      color='forest'
                      size='md'
                      disabled={!invitee || busy}
                      loading={busy}
                      onClick={() => {
                        if (!invitee) return;
                        run(async () => {
                          await groupAdapter.invite(group.id, Number(invitee));
                          setInvitee(null);
                        }, 'Could not send invite');
                      }}
                    >
                      Invite
                    </Button>
                  </Group>
                )}
              </Stack>
            )}
          </Tabs.Panel>

        {canManage && (
          <Tabs.Panel value='settings' pt='md'>
            <GroupSettingsPanel
              key={`${group.id}-${group.name}-${group.privacy}-${group.owner?.id}`}
              group={group}
              members={members}
              hostPlayer={hostPlayer}
              onUpdated={refresh}
            />
          </Tabs.Panel>
        )}
      </Tabs>
      )}

      <Modal opened={confirm?.kind === 'leave'} onClose={() => setConfirm(null)} title='Leave this group?' centered>
        <Text size='sm' mb='md'>
          You can join again later if the group is public, or request to join if it is private.
        </Text>
        <Stack gap='sm'>
          <Button
            color='red'
            size='md'
            loading={busy}
            onClick={() =>
              run(async () => {
                await groupAdapter.leave(group.id);
                setConfirm(null);
                navigate('/groups');
              }, 'Could not leave group')
            }
          >
            Leave group
          </Button>
          <Button variant='light' color='gray' size='md' onClick={() => setConfirm(null)}>
            Stay
          </Button>
        </Stack>
      </Modal>
      <Modal
        opened={confirm?.kind === 'remove'}
        onClose={() => setConfirm(null)}
        title='Remove member?'
        centered
      >
        <Text size='sm' mb='md'>
          {confirm?.kind === 'remove' ? confirm.name : ''} will lose access to this group.
        </Text>
        <Stack gap='sm'>
          <Button
            color='red'
            size='md'
            loading={busy}
            onClick={() => {
              if (confirm?.kind !== 'remove') return;
              const playerId = confirm.playerId;
              run(async () => {
                await groupAdapter.removeMember(group.id, playerId);
                setConfirm(null);
              }, 'Could not remove member');
            }}
          >
            Remove
          </Button>
          <Button variant='light' color='gray' size='md' onClick={() => setConfirm(null)}>
            Cancel
          </Button>
        </Stack>
      </Modal>
      <Modal
        opened={confirm?.kind === 'cancelInvite'}
        onClose={() => setConfirm(null)}
        title='Cancel invitation?'
        centered
      >
        <Text size='sm' mb='md'>
          {confirm?.kind === 'cancelInvite' ? confirm.name : ''} will no longer see this invite.
        </Text>
        <Stack gap='sm'>
          <Button
            color='red'
            size='md'
            loading={busy}
            onClick={() => {
              if (confirm?.kind !== 'cancelInvite') return;
              const invitationId = confirm.invitationId;
              run(async () => {
                await groupAdapter.cancelInvitation(group.id, invitationId);
                setConfirm(null);
              }, 'Could not cancel invitation');
            }}
          >
            Cancel invitation
          </Button>
          <Button variant='light' color='gray' size='md' onClick={() => setConfirm(null)}>
            Keep invite
          </Button>
        </Stack>
      </Modal>
    </Container>
  );
}
