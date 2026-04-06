import { useState } from 'react';
import {
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Title,
  Stack,
  Text,
  Center,
  Divider,
  Box,
  SegmentedControl,
  Group,
  Alert,
} from '@mantine/core';
import { useMantineColorScheme } from '@mantine/core';
import { FiSun, FiMoon, FiMonitor, FiCheck, FiAlertCircle } from 'react-icons/fi';
import { setColorScheme as persistColorScheme } from '../../adapters/storage/localStorageAdapter';
import type { Player, UpdateProfileRequest, ChangePasswordRequest } from '../../types';

interface ProfileProps {
  player: Player;
  onUpdateProfile: (data: UpdateProfileRequest) => Promise<void>;
  onChangePassword: (data: ChangePasswordRequest) => Promise<void>;
}

function Profile({ player, onUpdateProfile, onChangePassword }: ProfileProps) {
  // Personal info
  const [name, setName] = useState(player.name);
  const [phone, setPhone] = useState(player.phone);
  const [email, setEmail] = useState(player.email);
  const [username, setUsername] = useState(player.username);
  const [savingProfile, setSavingProfile] = useState(false);
  const [profileMessage, setProfileMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Password
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [savingPassword, setSavingPassword] = useState(false);
  const [passwordMessage, setPasswordMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Theme
  const { colorScheme, setColorScheme } = useMantineColorScheme();

  const handleSaveProfile = async () => {
    setSavingProfile(true);
    setProfileMessage(null);
    try {
      await onUpdateProfile({ name, phone, email, username });
      setProfileMessage({ type: 'success', text: 'Profile updated successfully' });
    } catch {
      setProfileMessage({ type: 'error', text: 'Failed to update profile. Please try again.' });
    } finally {
      setSavingProfile(false);
    }
  };

  const handleChangePassword = async () => {
    if (newPassword !== confirmPassword) {
      setPasswordMessage({ type: 'error', text: 'New passwords do not match' });
      return;
    }
    setSavingPassword(true);
    setPasswordMessage(null);
    try {
      await onChangePassword({
        current_password: currentPassword,
        new_password: newPassword,
        password_confirmation: confirmPassword,
      });
      setPasswordMessage({ type: 'success', text: 'Password changed successfully' });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch {
      setPasswordMessage({ type: 'error', text: 'Failed to change password. Check your current password.' });
    } finally {
      setSavingPassword(false);
    }
  };

  const handleColorSchemeChange = (value: string) => {
    const scheme = value as 'light' | 'dark' | 'auto';
    setColorScheme(scheme);
    persistColorScheme(scheme);
  };

  const themeData = [
    { label: 'Light', value: 'light' },
    { label: 'Dark', value: 'dark' },
    { label: 'System', value: 'auto' },
  ];

  return (
    <Center p='md' style={{ minHeight: 'calc(100vh - 64px)' }}>
      <Paper shadow='lg' p='xl' maw={520} w='100%'>
        <Box mb='xl'>
          <Title order={2} style={{ color: 'var(--ff-heading)' }} ta='center'>
            Profile & Settings
          </Title>
          <Text c='dimmed' size='sm' ta='center' mt={4}>
            Manage your account and preferences
          </Text>
        </Box>

        {/* Personal Info Section */}
        <Stack gap='md'>
          <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }}>Personal Info</Text>

          {profileMessage && (
            <Alert
              color={profileMessage.type === 'success' ? 'green' : 'red'}
              icon={profileMessage.type === 'success' ? <FiCheck /> : <FiAlertCircle />}
              withCloseButton
              onClose={() => setProfileMessage(null)}
            >
              {profileMessage.text}
            </Alert>
          )}

          <TextInput
            label='Full Name'
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <TextInput
            label='Phone'
            type='tel'
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
          />
          <TextInput
            label='Email'
            type='email'
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <TextInput
            label='Username'
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <Button
            color='forest'
            size='md'
            fullWidth
            onClick={handleSaveProfile}
            loading={savingProfile}
          >
            Save Profile
          </Button>
        </Stack>

        <Divider my='xl' color='var(--ff-border)' />

        {/* Change Password Section */}
        <Stack gap='md'>
          <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }}>Change Password</Text>

          {passwordMessage && (
            <Alert
              color={passwordMessage.type === 'success' ? 'green' : 'red'}
              icon={passwordMessage.type === 'success' ? <FiCheck /> : <FiAlertCircle />}
              withCloseButton
              onClose={() => setPasswordMessage(null)}
            >
              {passwordMessage.text}
            </Alert>
          )}

          <PasswordInput
            label='Current Password'
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
          />
          <PasswordInput
            label='New Password'
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
          />
          <PasswordInput
            label='Confirm New Password'
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
          />
          <Button
            color='forest'
            size='md'
            fullWidth
            onClick={handleChangePassword}
            loading={savingPassword}
            disabled={!currentPassword || !newPassword || !confirmPassword}
          >
            Change Password
          </Button>
        </Stack>

        <Divider my='xl' color='var(--ff-border)' />

        {/* Appearance Section */}
        <Stack gap='md'>
          <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }}>Appearance</Text>
          <Text size='sm' c='dimmed'>Choose your preferred color scheme</Text>

          <Group justify='center'>
            <SegmentedControl
              value={colorScheme}
              onChange={handleColorSchemeChange}
              data={themeData}
              color='forest'
            />
          </Group>

          <Group justify='center' gap='lg'>
            <Group gap={6}>
              <FiSun size={14} style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed'>Light</Text>
            </Group>
            <Group gap={6}>
              <FiMoon size={14} style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed'>Dark</Text>
            </Group>
            <Group gap={6}>
              <FiMonitor size={14} style={{ color: 'var(--ff-icon-primary)' }} />
              <Text size='xs' c='dimmed'>System</Text>
            </Group>
          </Group>
        </Stack>
      </Paper>
    </Center>
  );
}

export default Profile;
