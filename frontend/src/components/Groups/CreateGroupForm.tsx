import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Button,
  Container,
  Radio,
  Stack,
  Text,
  TextInput,
  Textarea,
  Title,
} from '@mantine/core';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';

export default function CreateGroupForm() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [privacy, setPrivacy] = useState<'public' | 'private'>('public');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const submit = async () => {
    setSaving(true);
    setError('');
    try {
      const created = await groupAdapter.create(name, description, privacy);
      navigate(`/groups/${created.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not create group');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Container size='sm' py='lg'>
      <Title order={2} mb='md'>
        Create Group
      </Title>
      <Stack gap='md'>
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
        {error && (
          <Text c='red' size='sm'>
            {error}
          </Text>
        )}
        <Button color='forest' onClick={submit} loading={saving} disabled={!name.trim()}>
          Create Group
        </Button>
      </Stack>
    </Container>
  );
}
