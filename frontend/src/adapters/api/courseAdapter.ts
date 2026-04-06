import type { Course } from '../../domain/course/types';
import type { CoursePort } from '../../ports/coursePort';
import { endpoints, request } from './httpClient';

export const courseAdapter: CoursePort = {
  getAll(): Promise<Course[]> {
    return request<Course[]>(endpoints.courses);
  },

  search(query: string): Promise<Course[]> {
    return request<Course[]>(
      `${endpoints.courses}/search?q=${encodeURIComponent(query)}`,
    );
  },

  findOrCreate(course: Omit<Course, 'id'>): Promise<Course> {
    return request<Course>(endpoints.courses, {
      method: 'POST',
      body: JSON.stringify(course),
    });
  },
};
