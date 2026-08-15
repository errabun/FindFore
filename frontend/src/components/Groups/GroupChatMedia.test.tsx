import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import GroupChatAttachmentSelector from './GroupChatMedia';
import { renderWithProviders } from '../../test/render';

vi.mock('stream-chat-react', () => ({
  AttachmentSelector: ({
    attachmentSelectorActionSet,
  }: {
    attachmentSelectorActionSet: Array<{ ActionButton: () => ReactNode }>;
  }) => (
    <div>
      {attachmentSelectorActionSet.map((action, i) => (
        <action.ActionButton key={i} />
      ))}
    </div>
  ),
  ContextMenuButton: ({ children }: { children: ReactNode }) => <button type='button'>{children}</button>,
  DefaultAttachmentSelectorComponents: {
    File: () => <button type='button'>File</button>,
  },
  FileInput: () => <input type='file' />,
  defaultAttachmentSelectorActionSet: [{ type: 'selectCommand', ActionButton: () => <button type='button'>Commands</button> }],
  useContextMenuContext: () => ({ closeMenu: vi.fn() }),
  useMessageComposerController: () => ({
    attachmentManager: { uploadFiles: vi.fn() },
    channel: { getConfig: () => ({ commands: [{ name: 'giphy' }] }) },
    textComposer: { setCommand: vi.fn() },
  }),
  useMessageComposerContext: () => ({ textareaRef: { current: null } }),
}));

describe('GroupChatAttachmentSelector', () => {
  it('offers photo, video, gif, and file actions', () => {
    renderWithProviders(<GroupChatAttachmentSelector />);
    expect(screen.getByRole('button', { name: 'Photo' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Video' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'GIF' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'File' })).toBeInTheDocument();
  });
});
