import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Burger, Anchor, Group, Drawer, Stack, UnstyledButton, Button } from '@mantine/core';
import { GiGolfTee } from 'react-icons/gi';
import { FiLogOut, FiUser } from 'react-icons/fi';

interface HeaderProps {
  screenWidth: number;
  isLoggedIn: boolean;
  onLogout: () => void;
}

const Header = ({ screenWidth, isLoggedIn, onLogout }: HeaderProps) => {
  const [mobileNav, setMobileNav] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    onLogout();
    setMobileNav(false);
    navigate('/login');
  };

  const isActive = (path: string) =>
    location.pathname === path || (path !== '/' && location.pathname.startsWith(`${path}/`));

  const navLinkStyle = (path: string): React.CSSProperties => ({
    color: 'var(--mantine-color-white)',
    fontSize: '1.1rem',
    fontWeight: 500,
    borderBottom: isActive(path) ? '2px solid var(--ff-accent)' : '2px solid transparent',
    height: 64,
    display: 'flex',
    alignItems: 'center',
    textDecoration: 'none',
    transition: 'border-color 0.2s ease',
  });

  const mobileLinkStyle: React.CSSProperties = {
    color: 'var(--mantine-color-white)',
    fontSize: '1.25rem',
    fontWeight: 500,
    textDecoration: 'none',
    padding: '0.75rem 0',
  };

  return (
    <header
      style={{
        background: 'var(--ff-header-bg)',
        height: 64,
        width: '100%',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        zIndex: 999,
        borderBottom: '1px solid rgba(255,255,255,0.1)',
      }}
    >
      <nav
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          height: 64,
          width: '100%',
          maxWidth: 1600,
          marginRight: 'auto',
          marginLeft: 'auto',
          paddingRight: screenWidth <= 991 ? 20 : 40,
          paddingLeft: screenWidth <= 991 ? 20 : 40,
          alignItems: 'center',
        }}
      >
        {screenWidth <= 1024 && (
          <Burger
            opened={mobileNav}
            onClick={() => setMobileNav(!mobileNav)}
            color='white'
            data-cy='ham-menu'
            size='sm'
          />
        )}

        <Anchor
          component={Link}
          to='/dashboard'
          onClick={() => setMobileNav(false)}
          underline='never'
          style={{
            color: 'var(--mantine-color-white)',
            fontSize: screenWidth <= 489 ? '1.4rem' : screenWidth <= 960 ? '1.6rem' : '1.8rem',
            fontWeight: 700,
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            textDecoration: 'none',
            letterSpacing: '-0.02em',
          }}
        >
          <GiGolfTee data-cy='logo' style={{ color: 'var(--ff-accent)', fontSize: '1.4em' }} />
          ForeFinder
        </Anchor>

        {screenWidth > 1024 ? (
          <Group gap='xl'>
            <Anchor
              component={Link}
              to='/dashboard'
              data-cy='dashboard-link'
              underline='never'
              style={navLinkStyle('/dashboard')}
            >
              Dashboard
            </Anchor>
            <Anchor
              component={Link}
              to='/event-form'
              data-cy='form-link'
              underline='never'
              style={navLinkStyle('/event-form')}
            >
              Create Tee Time
            </Anchor>
            {isLoggedIn && (
              <Anchor
                component={Link}
                to='/groups'
                data-cy='groups-link'
                underline='never'
                style={navLinkStyle('/groups')}
              >
                Groups
              </Anchor>
            )}
            {isLoggedIn && (
              <>
                <Anchor
                  component={Link}
                  to='/profile'
                  data-cy='profile-link'
                  underline='never'
                  style={navLinkStyle('/profile')}
                >
                  <FiUser size={16} style={{ marginRight: 6 }} />
                  Profile
                </Anchor>
                <Button
                  variant='subtle'
                  color='sand'
                  size='sm'
                  leftSection={<FiLogOut size={16} />}
                  onClick={handleLogout}
                  data-cy='logout-btn'
                  styles={{
                    root: {
                      color: 'var(--ff-accent)',
                      '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' },
                    },
                  }}
                >
                  Log Out
                </Button>
              </>
            )}
          </Group>
        ) : (
          <Drawer
            opened={mobileNav}
            onClose={() => setMobileNav(false)}
            size='280'
            position='left'
            data-cy='nav-menu'
            withCloseButton={false}
            overlayProps={{ backgroundOpacity: 0.4 }}
            styles={{
              content: {
                background: 'var(--ff-header-bg)',
              },
              body: {
                padding: '2rem 1.5rem',
              },
            }}
          >
            <Stack gap='xs' mt='md'>
              <UnstyledButton
                component={Link}
                to='/dashboard'
                data-cy='dashboard-link'
                onClick={() => setMobileNav(false)}
                style={mobileLinkStyle}
              >
                Dashboard
              </UnstyledButton>
              <UnstyledButton
                component={Link}
                to='/event-form'
                data-cy='form-link'
                onClick={() => setMobileNav(false)}
                style={mobileLinkStyle}
              >
                Create Tee Time
              </UnstyledButton>
              {isLoggedIn && (
                <UnstyledButton
                  component={Link}
                  to='/groups'
                  data-cy='groups-link'
                  onClick={() => setMobileNav(false)}
                  style={mobileLinkStyle}
                >
                  Groups
                </UnstyledButton>
              )}
              {screenWidth <= 1024 && (
                <UnstyledButton
                  component={Link}
                  to='/community'
                  data-cy='community-link'
                  onClick={() => setMobileNav(false)}
                  style={mobileLinkStyle}
                >
                  My Community
                </UnstyledButton>
              )}
              {isLoggedIn && (
                <>
                  <UnstyledButton
                    component={Link}
                    to='/profile'
                    data-cy='profile-link'
                    onClick={() => setMobileNav(false)}
                    style={{
                      ...mobileLinkStyle,
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                    }}
                  >
                    <FiUser size={18} />
                    Profile
                  </UnstyledButton>
                  <UnstyledButton
                    data-cy='logout-btn'
                    onClick={handleLogout}
                    style={{
                      ...mobileLinkStyle,
                      color: 'var(--ff-accent)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      marginTop: '1rem',
                      borderTop: '1px solid rgba(255,255,255,0.1)',
                      paddingTop: '1.5rem',
                    }}
                  >
                    <FiLogOut size={18} />
                    Log Out
                  </UnstyledButton>
                </>
              )}
            </Stack>
          </Drawer>
        )}
      </nav>
    </header>
  );
};

export default Header;
