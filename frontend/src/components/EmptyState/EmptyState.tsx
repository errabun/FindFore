import { Stack, ThemeIcon, Text, Button } from '@mantine/core';
import { Link } from 'react-router-dom';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  icon: ReactNode;
  title: string;
  description: string;
  actionLabel?: string;
  onAction?: () => void;
  actionHref?: string;
}

const EmptyState = ({ icon, title, description, actionLabel, onAction, actionHref }: EmptyStateProps) => {
  return (
    <Stack align='center' gap='md' py='xl'>
      <ThemeIcon size='xl' radius='xl' variant='light' color='forest'>
        {icon}
      </ThemeIcon>
      <div style={{ textAlign: 'center' }}>
        <Text fw={600} size='sm' mb={4}>{title}</Text>
        <Text c='dimmed' size='sm'>{description}</Text>
      </div>
      {actionLabel && actionHref && (
        <Button
          variant='light'
          color='forest'
          size='sm'
          component={Link}
          to={actionHref}
        >
          {actionLabel}
        </Button>
      )}
      {actionLabel && onAction && !actionHref && (
        <Button
          variant='light'
          color='forest'
          size='sm'
          onClick={onAction}
        >
          {actionLabel}
        </Button>
      )}
    </Stack>
  );
};

export default EmptyState;
