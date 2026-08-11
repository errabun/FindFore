export interface Course {
  id: number;
  name: string;
  street: string;
  city: string;
  state: string;
  zip_code: string;
  phone: string;
  cost: string;
  country?: string;
  latitude?: number | null;
  longitude?: number | null;
  timezone?: string;
  provider?: string;
  external_id?: string;
}
