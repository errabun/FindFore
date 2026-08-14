export interface Event {
  id: number;
  course_name: string;
  date: string;
  tee_time: string;
  open_spots: number;
  number_of_holes: string;
  private: boolean;
  host_name: string;
  host_id: number;
  accepted: number[];
  declined: number[];
  pending: number[];
  closed: number[];
  remaining_spots: number;
  group_id?: number;
  group_name?: string;
}
