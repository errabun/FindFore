import type { Course } from '../domain/course/types';

export interface CoursePort {
  getAll(): Promise<Course[]>;
  search(query: string): Promise<Course[]>;
  findOrCreate(course: Omit<Course, 'id'>): Promise<Course>;
}
