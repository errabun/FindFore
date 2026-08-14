import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Alert,
  Button,
  Modal,
  Radio,
  Select,
  Stack,
  Text,
  TextInput,
  Textarea,
} from '@mantine/core';
import { FiAlertCircle } from 'react-icons/fi';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';
import type { GroupMember, GroupSummary } from '../../domain/group/types';

interface GroupSettingsPanelProps {
  group: GroupSummary;
  members: GroupMember[];
  hostPlayer: number;
  onUpdated: () => Promise<unknown>;
}

export default function GroupSettingsPanel({
  group,
  members,
  hostPlayer,
  onUpdated,
}: GroupSettingsPanelProps) {
  const navigate = useNavigate();
  const isOwner = group.viewer_membership?.role === 'owner';
  const [name, setName] = useState(group.name);
  const [description, setDescription] = useState(group.description);
  const [privacy, setPrivacy] = useState<'public' | 'private'>(group.privacy);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [transferTo, setTransferTo] = useState<string | null>(null);
  const [confirm, setConfirm] = useState<'transfer' | 'delete' | null>(null);

  const transferOptions = members
    .filter((m) => m.player_id !== hostPlayer && m.status === 'active' && m.role !== 'owner')
    .map((m) => ({ value: String(m.player_id), label: m.player_name || `Player ${m.player_id}` }));

  const save = async () => {
    setSaving(true);
    setError('');
    try {
      await groupAdapter.update(group.id, name, description, privacy);
      await onUpdated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not update group');
    } finally {
      setSaving(false);
    }
  };

  const transfer = async () => {
    if (!transferTo) return;
    setSaving(true);
    setError('');
    try {
      await groupAdapter.transferOwnership(group.id, Number(transferTo));
      setConfirm(null);
      await onUpdated();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not transfer ownership');
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    setSaving(true);
    setError('');
    try {
      await groupAdapter.delete(group.id);
      navigate('/groups');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not delete group');
      setSaving(false);
    }
  };

  return (
    <Stack gap='md'>
      {error && (
        <Alert color='red' icon={<FiAlertCircle />}>
          {error}
        </Alert>
      )}
      <TextInput
        label='Name'
        required
        maxLength={100}
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
      />
      <Textarea
        label='Description'
        maxLength={1000}
        minRows={3}
        value={description}
        onChange={(e) => setDescription(e.currentTarget.value)}
      />
      <Radio.Group
        label='Privacy'
        value={privacy}
        onChange={(v) => setPrivacy(v as 'public' | 'private')}
      >
        <Stack gap='xs' mt='xs'>
          <Radio value='public' label='Public — anyone can join' />
          <Radio value='private' label='Private — members must be approved' />
        </Stack>
      </Radio.Group>
      <Button
        color='forest'
        onClick={save}
        loading={saving}
        disabled={!name.trim()}
        size='md'
      >
        Save changes
      </Button>

      {isOwner && (
        <>
          <Text fw={600} size='sm' mt='sm'>
            Ownership
          </Text>
          <Text size='sm' c='dimmed'>
            Transfer ownership to another member if you want to leave this group.
          </Text>
          {transferOptions.length === 0 ? (
            <Text size='sm' c='dimmed'>
              Invite another golfer before you can transfer ownership.
            </Text>
          ) : (
            <Select
              label='New owner'
              placeholder='Choose a member'
              data={transferOptions}
              value={transferTo}
              onChange={setTransferTo}
              searchable
            />
          )}
          <Button
            variant='light'
            color='forest'
            disabled={!transferTo}
            onClick={() => setConfirm('transfer')}
            size='md'
          >
            Transfer ownership
          </Button>
          <Button variant='light' color='red' onClick={() => setConfirm('delete')} size='md'>
            Delete group
          </Button>
        </>
      )}

      <Modal
        opened={confirm === 'transfer'}
        onClose={() => setConfirm(null)}
        title='Transfer ownership?'
        centered
      >
        <Text size='sm' mb='md'>
          You will become a member. The new owner can manage this group.
        </Text>
        <GroupButtons
          confirmLabel='Transfer'
          onCancel={() => setConfirm(null)}
          onConfirm={transfer}
          loading={saving}
        />
      </Modal>
      <Modal
        opened={confirm === 'delete'}
        onClose={() => setConfirm(null)}
        title='Delete this group?'
        centered
      >
        <Text size='sm' mb='md'>
          This removes the group, members, and outstanding invitations. This cannot be undone.
        </Text>
        <GroupButtons
          confirmLabel='Delete group'
          confirmColor='red'
          onCancel={() => setConfirm(null)}
          onConfirm={remove}
          loading={saving}
        />
      </Modal>
    </Stack>
  );
}

function GroupButtons({
  confirmLabel,
  confirmColor = 'forest',
  onCancel,
  onConfirm,
  loading,
}: {
  confirmLabel: string;
  confirmColor?: string;
  onCancel: () => void;
  onConfirm: () => void;
  loading: boolean;
}) {
  return (
    <Stack gap='sm'>
      <Button color={confirmColor} onClick={onConfirm} loading={loading} size='md'>
        {confirmLabel}
      </Button>
      <Button variant='light' color='gray' onClick={onCancel} disabled={loading} size='md'>
        Cancel
      </Button>
    </Stack>
  );
}
