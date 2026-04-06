import { useState, useEffect } from 'react';
import type { Course } from '../domain/course/types';
import { courseAdapter } from '../adapters/api/courseAdapter';

export function useCourses() {
  const [courses, setCourses] = useState<Course[]>([]);

  useEffect(() => {
    courseAdapter.getAll().then(setCourses);
  }, []);

  return { courses };
}
