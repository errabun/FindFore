import { useState, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { courseAdapter } from '../../adapters/api/courseAdapter';
import { teeTimeAdapter } from '../../adapters/api/teeTimeAdapter';
import {
  Paper,
  Select,
  Autocomplete,
  Button,
  Title,
  Stack,
  Radio,
  Checkbox,
  Group,
  Text,
  SimpleGrid,
  Center,
  Box,
  Loader,
} from '@mantine/core';
import { DateInput, TimeInput } from '@mantine/dates';
import dayjs from 'dayjs';
import type { Event, Course, Friend } from '../../types';

interface EditTeeTimeProps {
  event: Event;
  friends: Friend[];
  refreshEvents: () => void;
}

function EditTeeTime({ event, friends, refreshEvents }: EditTeeTimeProps) {
  const navigate = useNavigate();
  const tomorrowDate = dayjs().add(1, 'day');

  const [date, setDate] = useState<string | null>(dayjs(event.date).format('YYYY-MM-DD'));
  const [teeTime, setTeeTime] = useState(event.tee_time);
  const [openSpots, setOpenSpots] = useState<string | null>(String(event.open_spots));
  const [numHoles, setNumHoles] = useState(event.number_of_holes);
  const [isPrivate, setIsPrivate] = useState(event.private);
  const [saving, setSaving] = useState(false);

  // Friends invite state — pre-check friends already invited (pending/accepted/declined/closed)
  const alreadyInvited = new Set([
    ...event.pending,
    ...event.accepted,
    ...event.declined,
    ...event.closed,
  ]);
  const [selectedFriends, setSelectedFriends] = useState<number[]>([]);
  const [allFriends, setAllFriends] = useState(false);

  // Friends not yet invited
  const uninvitedFriends = friends.filter((f) => !alreadyInvited.has(f.id));

  const addFriendToInvite = (friendId: number, checked: boolean) => {
    if (checked) {
      setSelectedFriends([...selectedFriends, friendId]);
    } else {
      setSelectedFriends(selectedFriends.filter((f) => f !== friendId));
    }
  };

  const inviteAllFriends = (checked: boolean) => {
    if (checked) {
      setSelectedFriends(uninvitedFriends.map((f) => f.id));
      setAllFriends(true);
    } else {
      setSelectedFriends([]);
      setAllFriends(false);
    }
  };

  // Course search state
  const [courseQuery, setCourseQuery] = useState(event.course_name);
  const [courseResults, setCourseResults] = useState<Course[]>([]);
  const [selectedCourse, setSelectedCourse] = useState<Course | null>({
    id: 0,
    name: event.course_name,
    street: '',
    city: '',
    state: '',
    zip_code: '',
    phone: '',
    cost: '',
  });
  const [searching, setSearching] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const justSelectedRef = useRef(false);

  const courseLabel = (c: Course) =>
    `${c.name} — ${c.city}, ${c.state}`;

  const handleCourseSearch = useCallback((value: string) => {
    setCourseQuery(value);

    if (justSelectedRef.current) {
      justSelectedRef.current = false;
      return;
    }

    setSelectedCourse(null);

    if (debounceRef.current) clearTimeout(debounceRef.current);

    if (value.length < 3) {
      setCourseResults([]);
      return;
    }

    debounceRef.current = setTimeout(() => {
      setSearching(true);
      courseAdapter.search(value)
        .then((results) => setCourseResults(results))
        .catch(() => setCourseResults([]))
        .finally(() => setSearching(false));
    }, 300);
  }, []);

  const handleCourseSelect = (value: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    setSearching(false);
    justSelectedRef.current = true;
    const idx = parseInt(value, 10);
    const course = courseResults[idx];
    if (course) {
      setSelectedCourse(course);
      setCourseQuery(courseLabel(course));
    }
  };

  const handleSave = async () => {
    if (!selectedCourse || !teeTime) return;

    setSaving(true);
    try {
      let courseId = selectedCourse.id;

      // If a new course was selected from search (id=0), save it first
      if (!courseId) {
        const saved = await courseAdapter.findOrCreate({
          name: selectedCourse.name,
          street: selectedCourse.street,
          city: selectedCourse.city,
          state: selectedCourse.state,
          zip_code: selectedCourse.zip_code || '',
          phone: selectedCourse.phone || '',
          cost: selectedCourse.cost || '',
        });
        courseId = saved.id;
      }

      const formattedDate = dayjs(date).format('YYYY-MM-DD');
      await teeTimeAdapter.updateEvent(
        event.id,
        courseId,
        formattedDate,
        teeTime,
        openSpots || '2',
        numHoles,
        isPrivate,
        selectedFriends,
      );

      refreshEvents();
      navigate('/dashboard');
    } catch {
      setSaving(false);
    }
  };

  const teeTimeHour = teeTime ? parseInt(teeTime.split(':')[0]) : null;
  const isValidTime = teeTime && teeTimeHour !== null && teeTimeHour >= 7 && teeTimeHour <= 17;

  return (
    <Center p='md' style={{ minHeight: 'calc(100vh - 64px)' }}>
      <Paper shadow='lg' p='xl' maw={520} w='100%'>
        <form onSubmit={(e) => e.preventDefault()}>
          <Box mb='xl'>
            <Title order={2} style={{ color: 'var(--ff-heading)' }} ta='center'>
              Edit Tee Time
            </Title>
            <Text c='dimmed' size='sm' ta='center' mt={4}>
              Update the details for your round
            </Text>
          </Box>

          <Stack gap='lg'>
            <Box>
              <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }} mb='xs'>When</Text>
              <Stack gap='sm'>
                <DateInput
                  label='Date'
                  value={date}
                  onChange={setDate}
                  minDate={tomorrowDate.toDate()}
                  required
                />
                <TimeInput
                  label='Tee Time (7am to 5pm)'
                  value={teeTime}
                  onChange={(e) => setTeeTime(e.target.value)}
                  required
                />
              </Stack>
            </Box>

            <Box>
              <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }} mb='xs'>Details</Text>
              <Stack gap='sm'>
                <Autocomplete
                  label='Golf Course'
                  placeholder='Search for a course...'
                  value={courseQuery}
                  onChange={handleCourseSearch}
                  onOptionSubmit={handleCourseSelect}
                  data={courseResults.map((c, i) => ({
                    value: String(i),
                    label: courseLabel(c),
                  }))}
                  rightSection={searching ? <Loader size='xs' /> : undefined}
                  required
                />
                <Select
                  label='Total Players (including you)'
                  value={openSpots}
                  onChange={setOpenSpots}
                  data={[
                    { value: '2', label: '2' },
                    { value: '3', label: '3' },
                    { value: '4', label: '4' },
                  ]}
                />
                <Radio.Group
                  label='Number of Holes'
                  value={numHoles}
                  onChange={setNumHoles}
                >
                  <Group mt='xs'>
                    <Radio value='18' label='18' color='forest' />
                    <Radio value='9' label='9' color='forest' />
                  </Group>
                </Radio.Group>
                <Radio.Group
                  label='Public or Private'
                  value={isPrivate ? 'private' : 'public'}
                  onChange={(val) => setIsPrivate(val === 'private')}
                >
                  <Group mt='xs'>
                    <Radio value='public' label='Public' color='forest' />
                    <Radio value='private' label='Private' color='forest' />
                  </Group>
                </Radio.Group>
              </Stack>
            </Box>

            {isPrivate && (
              <Box>
                <Text fw={600} size='sm' style={{ color: 'var(--ff-label)' }} mb='xs'>Invite Friends</Text>
                <Stack gap='sm'>
                  {!uninvitedFriends.length && (
                    <Text fs='italic' c='dimmed' size='sm'>
                      All your friends have already been invited.
                    </Text>
                  )}
                  {uninvitedFriends.length > 0 && (
                    <>
                      <SimpleGrid cols={2}>
                        {uninvitedFriends.map((friend) => (
                          <Checkbox
                            key={friend.id}
                            label={friend.name}
                            value={String(friend.id)}
                            checked={selectedFriends.includes(friend.id)}
                            onChange={(e) =>
                              addFriendToInvite(friend.id, e.currentTarget.checked)
                            }
                            disabled={allFriends}
                            color='forest'
                          />
                        ))}
                      </SimpleGrid>
                      <Checkbox
                        label='Invite All Friends'
                        checked={allFriends}
                        onChange={(e) =>
                          inviteAllFriends(e.currentTarget.checked)
                        }
                        color='forest'
                      />
                    </>
                  )}
                </Stack>
              </Box>
            )}

            <Group grow>
              <Button
                variant='light'
                color='forest'
                size='md'
                onClick={() => navigate('/dashboard')}
              >
                Cancel
              </Button>
              <Button
                color='forest'
                size='md'
                disabled={!selectedCourse || !isValidTime}
                loading={saving}
                onClick={handleSave}
              >
                Save Changes
              </Button>
            </Group>
          </Stack>
        </form>
      </Paper>
    </Center>
  );
}

export default EditTeeTime;
