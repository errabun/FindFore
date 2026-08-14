import { Link } from 'react-router-dom';
import { Modal, Text, Button, Stack, ThemeIcon } from '@mantine/core';
import { FiCheck, FiAlertTriangle } from 'react-icons/fi';

interface PostResultMessageProps {
  postError: boolean;
  errorMessage?: string;
  refreshEvents: () => void;
  successHref?: string;
  successLabel?: string;
}

function PostResultMessage({
  postError,
  errorMessage,
  refreshEvents,
  successHref = '/dashboard',
  successLabel = 'Back to Dashboard',
}: PostResultMessageProps) {
  return (
    <Modal
      opened
      onClose={!postError ? refreshEvents : () => {}}
      centered
      withCloseButton={false}
      overlayProps={{ backgroundOpacity: 0.3 }}
      radius='lg'
    >
      <Stack className='message' align='center' gap='md' py='lg'>
        {postError ? (
          <>
            <ThemeIcon size='xl' radius='xl' color='red' variant='light'>
              <FiAlertTriangle size={24} />
            </ThemeIcon>
            <Text ta='center'>
              {errorMessage ||
                "Sorry, we weren't able to send your event invitation. Please try again later."}
            </Text>
            <Button component={Link} to='/dashboard' color='forest' variant='light'>
              Back to Dashboard
            </Button>
          </>
        ) : (
          <>
            <ThemeIcon size='xl' radius='xl' color='green' variant='light'>
              <FiCheck size={24} />
            </ThemeIcon>
            <Text ta='center' fw={500}>
              Congrats, your tee time has been created!
            </Text>
            <Button component={Link} to={successHref} color='forest' onClick={refreshEvents}>
              {successLabel}
            </Button>
          </>
        )}
      </Stack>
    </Modal>
  );
}

export default PostResultMessage;
