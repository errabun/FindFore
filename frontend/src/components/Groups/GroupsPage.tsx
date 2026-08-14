import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Badge,
  Button,
  Card,
  Container,
  Group,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { FiPlus, FiUsers } from 'react-icons/fi';
import EmptyState from '../EmptyState/EmptyState';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { useGroups } from '../../hooks/useGroups';
import type { GroupInvitation, GroupSummary } from '../../domain/group/types';

interface GroupsPageProps {
  hostPlayer: number;
}

function GroupCard({ group }: { group: GroupSummary }) {
  return (
    <Card
      component={Link}
      to={`/groups/${group.id}`}
      padding='md'
      radius='md'
      withBorder
      style={{ textDecoration: 'none', color: 'inherit' }}
    >
      <Group justify='space-between' wrap='nowrap'>
        <div>
          <Text fw={600}>{group.name}</Text>
          <Text size='sm' c='dimmed'>
            {group.member_count} {group.member_count === 1 ? 'member' : 'members'}
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
  const { mine, discover, invitations, loading, refresh, setDiscover } = useGroups(hostPlayer);
  const [search, setSearch] = useState('');

  const onSearch = (value: string) => {
    setSearch(value);
    groupAdapter.listDiscover(value).then(setDiscover).catch(() => undefined);
  };

  const accept = (inv: GroupInvitation) => {
    groupAdapter.acceptInvitation(inv.id).then(() => refresh());
  };
  const decline = (inv: GroupInvitation) => {
    groupAdapter.declineInvitation(inv.id).then(() => refresh());
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
          size='sm'
        >
          New
        </Button>
      </Group>

      {invitations.length > 0 && (
        <Stack gap='sm' mb='xl'>
          <Text fw={600}>Invitations ({invitations.length})</Text>
          {invitations.map((inv) => (
            <Card key={inv.id} withBorder padding='md' radius='md'>
              <Text size='sm' mb='sm'>
                {inv.inviter_name} invited you to {inv.group_name}
              </Text>
              <Group>
                <Button size='xs' color='forest' onClick={() => accept(inv)}>
                  Accept
                </Button>
                <Button size='xs' variant='light' color='gray' onClick={() => decline(inv)}>
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
      {mine.length === 0 && !loading ? (
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
        placeholder='Search groups...'
        value={search}
        onChange={(e) => onSearch(e.currentTarget.value)}
        mb='sm'
      />
      <Stack gap='sm'>
        {discover.map((g) => (
          <GroupCard key={g.id} group={g} />
        ))}
      </Stack>
    </Container>
  );
}
