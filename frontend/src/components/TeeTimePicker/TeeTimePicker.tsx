import { Stack } from '@mantine/core';
import { TimePicker } from '@mantine/dates';
import { FiClock } from 'react-icons/fi';

export const TEE_TIME_MIN = '05:00';
export const TEE_TIME_MAX = '20:00';

export function normalizeTeeTimeValue(value: string): string {
  if (!value) return '';
  const [hours, minutes] = value.split(':');
  if (!hours || minutes === undefined) return value;
  return `${hours.padStart(2, '0')}:${minutes.slice(0, 2)}`;
}

interface TeeTimePickerProps {
  value: string;
  onChange: (value: string) => void;
}

function TeeTimePicker({ value, onChange }: TeeTimePickerProps) {
  const normalized = normalizeTeeTimeValue(value);

  return (
    <Stack gap='xs'>
      <TimePicker
        label='Tee time'
        description='Between 5:00 AM and 8:00 PM'
        format='12h'
        withDropdown
        min={TEE_TIME_MIN}
        max={TEE_TIME_MAX}
        minutesStep={10}
        maxDropdownContentHeight={280}
        size='md'
        required
        value={normalized}
        onChange={onChange}
        hoursInputLabel='Hours'
        minutesInputLabel='Minutes'
        amPmInputLabel='AM/PM'
        leftSection={<FiClock size={16} aria-hidden style={{ pointerEvents: 'none' }} />}
      />
    </Stack>
  );
}

export default TeeTimePicker;
