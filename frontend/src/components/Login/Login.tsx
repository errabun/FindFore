import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Paper, TextInput, PasswordInput, Button, Title, Stack, Text, Center, Box } from '@mantine/core';
import { GiGolfTee } from 'react-icons/gi';

interface LoginProps {
  validateLogin: (login: string, password: string) => void;
  loginError?: string;
  clearLoginError?: () => void;
}

function Login({ validateLogin, loginError, clearLoginError }: LoginProps) {
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');

  return (
    <Center style={{ minHeight: 'calc(100vh - 64px)' }} p='md'>
      <Paper shadow='lg' p='xl' maw={420} w='100%'>
        <form onSubmit={(e) => e.preventDefault()}>
          <Stack align='center' gap='xs' mb='xl'>
            <Box style={{ color: 'var(--ff-link)', fontSize: '2.5rem' }}>
              <GiGolfTee />
            </Box>
            <Title order={2} ta='center' style={{ color: 'var(--ff-heading)' }}>
              Welcome back
            </Title>
            <Text c='dimmed' size='sm'>
              Sign in to find your next round
            </Text>
          </Stack>

          <Stack gap='md'>
            {loginError && (
              <Text c='red.6' size='sm' ta='center'>
                {loginError}
              </Text>
            )}
            <TextInput
              label='Email or Username'
              id='login'
              name='login'
              value={login}
              onChange={(e) => {
                if (loginError) {
                  clearLoginError?.();
                }
                setLogin(e.target.value);
              }}
              required
            />
            <PasswordInput
              label='Password'
              id='password'
              name='password'
              value={password}
              onChange={(e) => {
                if (loginError) {
                  clearLoginError?.();
                }
                setPassword(e.target.value);
              }}
              required
            />
            <Button
              color='forest'
              size='md'
              fullWidth
              onClick={() => validateLogin(login, password)}
              className='form-submit'
              mt='sm'
            >
              Sign In
            </Button>
          </Stack>

          <Text ta='center' size='sm' c='dimmed' mt='lg'>
            Don't have an account?{' '}
            <Text component={Link} to='/create-profile' style={{ color: 'var(--ff-link)', textDecoration: 'none' }} fw={600} inherit>
              Create one
            </Text>
          </Text>
        </form>
      </Paper>
    </Center>
  );
}

export default Login;
