export function formatTeeTime(time: string): string {
  let hours: string | number = time.split(':')[0];
  const minutes = time.split(':')[1];
  const period = parseInt(hours) > 11 ? 'PM' : 'AM';

  if (parseInt(hours) < 10) {
    hours = hours.slice(1, 2);
  } else if (parseInt(hours) > 12) {
    hours = parseInt(hours) - 12;
  }

  return `${hours}:${minutes} ${period}`;
}
