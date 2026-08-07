import { useState } from 'react';
import { authAdapter } from '../../adapters/api/authAdapter';
import { ApiError } from '../../adapters/api/httpClient';
import { Paper, TextInput, PasswordInput, Button, Title, Stack, Text, Center, Divider, Box } from '@mantine/core';
import { GiGolfTee } from 'react-icons/gi';
import { useNavigate } from 'react-router-dom';

function CreateProfile() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [userName, setUserName] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [formError, setFormError] = useState('');

  const submitProfile = () => {
    setFormError('');
    if (password !== passwordConfirm) {
      setFormError('Passwords do not match, please try again!');
      return;
    }
    if (password.length < 8) {
      setFormError('Password must be at least 8 characters');
      return;
    }
    authAdapter
      .createProfile(name, phone, email, userName, password, passwordConfirm)
      .then(() => navigate('/login', { replace: true }))
      .catch((err) => {
        if (err instanceof ApiError) {
          setFormError(err.message);
          return;
        }
        setFormError('Unable to create new profile, please try again!');
      });
  };

  return (
    <Center style={{ minHeight: 'calc(100vh - 64px)' }} p='md'>
      <Paper shadow='lg' p='xl' maw={420} w='100%'>
        <form onSubmit={(e) => e.preventDefault()}>
          <Stack align='center' gap='xs' mb='xl'>
            <Box style={{ color: 'var(--ff-link)', fontSize: '2.5rem' }}>
              <GiGolfTee />
            </Box>
            <Title order={2} ta='center' style={{ color: 'var(--ff-heading)' }}>
              Create your profile
            </Title>
            <Text c='dimmed' size='sm'>
              Join the ForeFinder community
            </Text>
          </Stack>

          <Stack gap='md'>
            {formError && (
              <Text c='red.6' size='sm' ta='center'>
                {formError}
              </Text>
            )}
            <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }}>Personal Info</Text>
            <TextInput
              label='Full Name'
              id='name'
              name='name'
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder='ex: John Doe'
              required
            />
            <TextInput
              label='Phone'
              type='tel'
              id='phone'
              name='phone'
              placeholder='123-456-7890'
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              required
            />
            <TextInput
              label='Email'
              type='email'
              id='email'
              name='email'
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder='ex: john.doe@example.com'
              required
            />

            <Divider my='xs' color='sand.2' />

            <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }}>Account</Text>
            <TextInput
              label='Username'
              id='userName'
              name='userName'
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
              placeholder='golfer1234'
              required
            />
            <PasswordInput
              label='Password'
              description='At least 8 characters'
              id='password'
              name='password'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete='new-password'
              required
            />
            <PasswordInput
              label='Confirm Password'
              id='passwordConfirm'
              name='passwordConfirm'
              value={passwordConfirm}
              onChange={(e) => setPasswordConfirm(e.target.value)}
              autoComplete='new-password'
              required
            />
            <Button
              color='forest'
              size='md'
              onClick={submitProfile}
              fullWidth
              mt='sm'
              className='form-submit'
            >
              Create Profile
            </Button>
            <Button
              variant='subtle'
              color='forest'
              size='sm'
              onClick={() => navigate('/login')}
              fullWidth
            >
              Back to Login
            </Button>
          </Stack>
        </form>
      </Paper>
    </Center>
  );
}

export default CreateProfile;
