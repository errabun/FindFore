import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import GroupChat from './GroupChat';
import { renderWithProviders } from '../../test/render';
import { groupAdapter } from '../../adapters/api/groupAdapter';
import { ApiError } from '../../adapters/api/httpClient';

vi.mock('stream-chat-react/dist/css/index.css', () => ({}));

vi.mock('stream-chat-react', () => ({
  Chat: ({ children }: { children: ReactNode }) => <div data-testid='stream-chat'>{children}</div>,
  Channel: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Window: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  ChannelHeader: () => <div>Channel header</div>,
  MessageList: () => <div>Messages</div>,
  MessageComposer: () => <div>Composer</div>,
  Thread: () => null,
  WithComponents: ({ children }: { children: ReactNode }) => <>{children}</>,
  useCreateChatClient: () => ({
    channel: () => ({}),
  }),
  AttachmentSelector: () => null,
  ContextMenuButton: ({ children }: { children: ReactNode }) => <button type='button'>{children}</button>,
  DefaultAttachmentSelectorComponents: { File: () => null },
  FileInput: () => null,
  defaultAttachmentSelectorActionSet: [],
  useContextMenuContext: () => ({ closeMenu: vi.fn() }),
  useMessageComposerController: () => ({
    attachmentManager: { uploadFiles: vi.fn() },
    channel: { getConfig: () => ({ commands: [] }) },
    textComposer: { setCommand: vi.fn() },
  }),
  useMessageComposerContext: () => ({ textareaRef: { current: null } }),
}));

vi.mock('../../adapters/api/groupAdapter', () => ({
  groupAdapter: {
    getChat: vi.fn(),
  },
}));

const mocked = vi.mocked(groupAdapter);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('GroupChat', () => {
  it('connects after loading a session', async () => {
    mocked.getChat.mockResolvedValue({
      api_key: 'pk',
      token: 'tok',
      channel_type: 'messaging',
      channel_id: 'group_10',
      user_id: '1',
      user_name: 'Eric',
    });

    renderWithProviders(<GroupChat groupId={10} />);
    expect(await screen.findByTestId('stream-chat')).toBeInTheDocument();
    expect(screen.getByText('Composer')).toBeInTheDocument();
    await waitFor(() => expect(mocked.getChat).toHaveBeenCalledWith(10));
  });

  it('shows an unavailable state when chat is not configured', async () => {
    mocked.getChat.mockRejectedValue(new ApiError(503, 'unavailable', 'Chat is not configured'));
    renderWithProviders(<GroupChat groupId={10} />);
    expect(await screen.findByText(/chat isn't available yet/i)).toBeInTheDocument();
    expect(screen.queryByTestId('stream-chat')).not.toBeInTheDocument();
  });
});
