import { useState } from 'react';
import { SegmentedControl } from '@mantine/core';

interface TypeSelectorProps {
  handleClick: (value: string) => void;
}

const TypeSelector = ({ handleClick }: TypeSelectorProps) => {
  const [value, setValue] = useState('all');

  return (
    <SegmentedControl
      value={value}
      onChange={(val) => {
        setValue(val);
        handleClick(val);
      }}
      size='xs'
      color='forest'
      radius='xl'
      data={[
        { label: 'All', value: 'all' },
        { label: 'Friends', value: 'friends' },
        { label: 'Public', value: 'public' },
        { label: 'Open', value: 'join' },
      ]}
    />
  );
};

export default TypeSelector;
