import { useRef, type ComponentType, type SVGProps } from 'react';
import { FiImage, FiVideo } from 'react-icons/fi';
import {
  AttachmentSelector,
  ContextMenuButton,
  DefaultAttachmentSelectorComponents,
  FileInput,
  defaultAttachmentSelectorActionSet,
  useContextMenuContext,
  useMessageComposerController,
  useMessageComposerContext,
} from 'stream-chat-react';
import type { AttachmentSelectorAction } from 'stream-chat-react';

const PHOTO_ACCEPT = 'image/jpeg,image/png,image/webp,image/heic,image/gif';
const VIDEO_ACCEPT = 'video/mp4,video/quicktime,video/webm';
const GIF_ACCEPT = 'image/gif';

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>;

function GifIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox='0 0 24 24' fill='none' stroke='currentColor' strokeWidth='2' {...props}>
      <rect x='3' y='5' width='18' height='14' rx='2' />
      <path d='M8 15V9h2.2a2 2 0 0 1 0 4H8' />
      <path d='M13 9v6M16 9v6' />
    </svg>
  );
}

function MediaUploadButton({
  accept,
  label,
  Icon,
}: {
  accept: string;
  label: string;
  Icon: IconComponent;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const { attachmentManager } = useMessageComposerController();
  const { closeMenu } = useContextMenuContext();

  return (
    <>
      <ContextMenuButton
        className='str-chat__attachment-selector-actions-menu__button'
        Icon={Icon}
        onClick={() => {
          inputRef.current?.click();
          closeMenu();
        }}
      >
        {label}
      </ContextMenuButton>
      <FileInput
        accept={accept}
        className='str-chat__file-input'
        multiple
        onFileChange={(files) => {
          void attachmentManager.uploadFiles(files);
        }}
        ref={inputRef}
        style={{ display: 'none' }}
      />
    </>
  );
}

function PhotoButton() {
  return <MediaUploadButton accept={PHOTO_ACCEPT} label='Photo' Icon={FiImage as IconComponent} />;
}

function VideoButton() {
  return <MediaUploadButton accept={VIDEO_ACCEPT} label='Video' Icon={FiVideo as IconComponent} />;
}

function GifButton() {
  const gifInputRef = useRef<HTMLInputElement>(null);
  const messageComposer = useMessageComposerController();
  const { textareaRef } = useMessageComposerContext();
  const { closeMenu } = useContextMenuContext();
  const giphy = messageComposer.channel.getConfig()?.commands?.find((command) => command.name === 'giphy');

  return (
    <>
      <ContextMenuButton
        className='str-chat__attachment-selector-actions-menu__button'
        Icon={GifIcon}
        onClick={() => {
          if (giphy) {
            messageComposer.textComposer.setCommand(giphy);
            closeMenu();
            requestAnimationFrame(() => textareaRef.current?.focus());
            return;
          }
          gifInputRef.current?.click();
          closeMenu();
        }}
      >
        GIF
      </ContextMenuButton>
      <FileInput
        accept={GIF_ACCEPT}
        className='str-chat__file-input'
        onFileChange={(files) => {
          void messageComposer.attachmentManager.uploadFiles(files);
        }}
        ref={gifInputRef}
        style={{ display: 'none' }}
      />
    </>
  );
}

const commandsAction = defaultAttachmentSelectorActionSet.find((action) => action.type === 'selectCommand');

const mediaActionSet: AttachmentSelectorAction[] = [
  { type: 'uploadFile', ActionButton: PhotoButton },
  { type: 'uploadFile', ActionButton: VideoButton },
  { type: 'giphy', ActionButton: GifButton },
  { type: 'uploadFile', ActionButton: DefaultAttachmentSelectorComponents.File },
  ...(commandsAction ? [commandsAction] : []),
];

export default function GroupChatAttachmentSelector() {
  return <AttachmentSelector attachmentSelectorActionSet={mediaActionSet} />;
}
