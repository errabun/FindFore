import { useEffect, useMemo, useState } from 'react';
import { Alert, Center, Loader, Text } from '@mantine/core';
import {
  Channel,
  ChannelHeader,
  Chat,
  MessageComposer,
  MessageList,
  Thread,
  Window,
  WithComponents,
  useCreateChatClient,
} from 'stream-chat-react';
import type { ReactionOptions } from 'stream-chat-react';
import 'stream-chat-react/dist/css/index.css';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';
import type { GroupChatSession } from '../../domain/group/types';
import GroupChatAttachmentSelector from './GroupChatMedia';
import './GroupChat.css';

const reaction = (emoji: string, name: string) => ({
  Component: () => <>{emoji}</>,
  name,
});

const golfReactions: ReactionOptions = {
  quick: {
    like: reaction('👍', 'Like'),
    love: reaction('❤️', 'Love'),
    haha: reaction('😂', 'Haha'),
    fire: reaction('🔥', 'Fire'),
    golf: reaction('⛳', 'Golf'),
    clap: reaction('👏', 'Clap'),
    party: reaction('🎉', 'Party'),
    muscle: reaction('💪', 'Strong'),
  },
  extended: {
    like: reaction('👍', 'Like'),
    love: reaction('❤️', 'Love'),
    haha: reaction('😂', 'Haha'),
    fire: reaction('🔥', 'Fire'),
    golf: reaction('⛳', 'Golf'),
    clap: reaction('👏', 'Clap'),
    party: reaction('🎉', 'Party'),
    muscle: reaction('💪', 'Strong'),
  },
};

interface GroupChatProps {
  groupId: number;
}

export default function GroupChat({ groupId }: GroupChatProps) {
  const [session, setSession] = useState<GroupChatSession | null>(null);
  const [error, setError] = useState('');
  const [unavailable, setUnavailable] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setSession(null);
    setError('');
    setUnavailable(false);
    groupAdapter
      .getChat(groupId)
      .then((next) => {
        if (!cancelled) setSession(next);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        if (err instanceof ApiError && (err.status === 503 || err.code === 'unavailable')) {
          setUnavailable(true);
          return;
        }
        setError(err instanceof ApiError ? err.message : 'Could not open group chat');
      });
    return () => {
      cancelled = true;
    };
  }, [groupId]);

  if (unavailable) {
    return (
      <Text size='sm' c='dimmed'>
        Chat isn't available yet. Activity is still the place to post for the group.
      </Text>
    );
  }
  if (error) {
    return (
      <Alert color='red' title='Chat unavailable'>
        {error}
      </Alert>
    );
  }
  if (!session) {
    return (
      <Center py='xl'>
        <Loader color='forest' />
      </Center>
    );
  }
  return <ConnectedGroupChat session={session} />;
}

function ConnectedGroupChat({ session }: { session: GroupChatSession }) {
  const client = useCreateChatClient({
    apiKey: session.api_key,
    tokenOrProvider: session.token,
    userData: { id: session.user_id, name: session.user_name },
  });
  const channel = useMemo(() => {
    if (!client) return undefined;
    return client.channel(session.channel_type, session.channel_id);
  }, [client, session.channel_id, session.channel_type]);

  if (!client || !channel) {
    return (
      <Center py='xl'>
        <Loader color='forest' />
      </Center>
    );
  }

  return (
    <div className='ff-group-chat'>
      <Chat client={client}>
        <WithComponents
          overrides={{
            reactionOptions: golfReactions,
            AttachmentSelector: GroupChatAttachmentSelector,
            ShareLocationDialog: () => null,
            StartRecordingAudioButton: () => null,
          }}
        >
          <Channel channel={channel}>
            <Window>
              <ChannelHeader />
              <MessageList />
              <MessageComposer />
            </Window>
            <Thread />
          </Channel>
        </WithComponents>
      </Chat>
    </div>
  );
}
