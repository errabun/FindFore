import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Alert,
  Badge,
  Button,
  Card,
  Container,
  Group,
  Skeleton,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { FiAlertCircle, FiPlus, FiSearch, FiUsers } from 'react-icons/fi';
import EmptyState from '../EmptyState/EmptyState';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';
import { useGroups } from '../../hooks/useGroups';
import type { GroupInvitation, GroupSummary } from '../../domain/group/types';

interface GroupsPageProps {
  hostPlayer: number;
}

function GroupCard({ group }: { group: GroupSummary }) {
  const joined = group.viewer_membership?.status === 'active';
  return (
    <Card
      component={Link}
      to={`/groups/${group.id}`}
      padding='md'
      radius='md'
      withBorder
      style={{ textDecoration: 'none', color: 'inherit', minHeight: 44 }}
    >
      <Group justify='space-between' wrap='nowrap'>
        <div>
          <Text fw={600}>{group.name}</Text>
          <Text size='sm' c='dimmed'>
            {group.member_count} {group.member_count === 1 ? 'member' : 'members'}
            {joined ? ' · Joined' : ''}
          </Text>
        </div>
        <Badge variant='light' color={group.privacy === 'public' ? 'teal' : 'gray'}>
          {group.privacy === 'public' ? 'Public' : 'Private'}
        </Badge>
      </Group>
    </Card>
  );
}

export default function GroupsPage({ hostPlayer }: GroupsPageProps) {
  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebouncedValue(search, 300);
  const {
    mine,
    discover,
    discoverHasMore,
    invitations,
    loading,
    error,
    refresh,
    loadMoreDiscover,
  } = useGroups(hostPlayer, debouncedSearch);
  const [actionError, setActionError] = useState('');
  const [busyId, setBusyId] = useState<number | null>(null);

  const accept = async (inv: GroupInvitation) => {
    setBusyId(inv.id);
    setActionError('');
    try {
      await groupAdapter.acceptInvitation(inv.id);
      await refresh();
    } catch (err) {
      setActionError(err instanceof ApiError ? err.message : 'Could not accept invitation');
    } finally {
      setBusyId(null);
    }
  };
  const decline = async (inv: GroupInvitation) => {
    setBusyId(inv.id);
    setActionError('');
    try {
      await groupAdapter.declineInvitation(inv.id);
      await refresh();
    } catch (err) {
      setActionError(err instanceof ApiError ? err.message : 'Could not decline invitation');
    } finally {
      setBusyId(null);
    }
  };

  return (
    <Container size='sm' py='lg'>
      <Group justify='space-between' mb='lg'>
        <Title order={2}>Groups</Title>
        <Button
          component={Link}
          to='/groups/new'
          color='forest'
          leftSection={<FiPlus />}
          size='md'
        >
          New
        </Button>
      </Group>

      {(error || actionError) && (
        <Alert color='red' icon={<FiAlertCircle />} mb='md'>
          {actionError || error}
        </Alert>
      )}

      {invitations.length > 0 && (
        <Stack gap='sm' mb='xl'>
          <Text fw={600}>Invitations ({invitations.length})</Text>
          {invitations.map((inv) => (
            <Card key={inv.id} withBorder padding='md' radius='md'>
              <Text size='sm' mb='sm'>
                {inv.inviter_name} invited you to {inv.group_name}
              </Text>
              <Group>
                <Button
                  size='sm'
                  color='forest'
                  loading={busyId === inv.id}
                  onClick={() => accept(inv)}
                >
                  Accept
                </Button>
                <Button
                  size='sm'
                  variant='light'
                  color='gray'
                  disabled={busyId === inv.id}
                  onClick={() => decline(inv)}
                >
                  Decline
                </Button>
              </Group>
            </Card>
          ))}
        </Stack>
      )}

      <Text fw={600} mb='sm'>
        My Groups
      </Text>
      {loading && mine.length === 0 ? (
        <Stack gap='sm' mb='xl'>
          <Skeleton height={64} radius='md' />
          <Skeleton height={64} radius='md' />
        </Stack>
      ) : mine.length === 0 ? (
        <EmptyState
          icon={<FiUsers />}
          title='No groups yet'
          description='Create a group or discover one to join.'
          actionLabel='Create group'
          actionHref='/groups/new'
        />
      ) : (
        <Stack gap='sm' mb='xl'>
          {mine.map((g) => (
            <GroupCard key={g.id} group={g} />
          ))}
        </Stack>
      )}

      <Text fw={600} mb='sm'>
        Discover Groups
      </Text>
      <TextInput
        placeholder='Search public groups'
        aria-label='Search public groups'
        leftSection={<FiSearch />}
        value={search}
        onChange={(e) => setSearch(e.currentTarget.value)}
        mb='sm'
      />
      {discover.length === 0 && !loading ? (
        <EmptyState
          icon={<FiUsers />}
          title={search ? 'No matching groups' : 'No public groups yet'}
          description={
            search
              ? 'Try a different name, or create your own group.'
              : 'Be the first to start a public group.'
          }
        />
      ) : (
        <Stack gap='sm'>
          {discover.map((g) => (
            <GroupCard key={g.id} group={g} />
          ))}
          {discoverHasMore && (
            <Button variant='light' color='forest' onClick={() => loadMoreDiscover()}>
              Load more
            </Button>
          )}
        </Stack>
      )}
    </Container>
  );
}
