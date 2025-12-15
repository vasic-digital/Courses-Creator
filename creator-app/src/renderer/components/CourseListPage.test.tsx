import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import CourseListPage from './CourseListPage';

// Mock the API
jest.mock('../services/api', () => ({
  publicAPI: {
    getCourses: jest.fn(),
  },
  courseAPI: {
    getCourses: jest.fn(),
  },
}));

const mockCourses = [
  {
    id: '1',
    title: 'Test Course 1',
    description: 'Description 1',
    created_at: '2023-01-01T00:00:00Z',
    updated_at: '2023-01-01T00:00:00Z',
    metadata: {
      author: 'Author 1',
      language: 'en',
      total_duration: 120,
    },
  },
  {
    id: '2',
    title: 'Test Course 2',
    description: 'Description 2',
    created_at: '2023-01-02T00:00:00Z',
    updated_at: '2023-01-02T00:00:00Z',
  },
];

describe('CourseListPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders loading state initially', () => {
    const { publicAPI } = require('../services/api');
    publicAPI.getCourses.mockImplementation(() => new Promise(() => {})); // Never resolves

    render(<CourseListPage isAuthenticated={false} />);

    expect(screen.getByText('Loading courses...')).toBeInTheDocument();
  });

  it('renders courses for authenticated user', async () => {
    const { courseAPI } = require('../services/api');
    courseAPI.getCourses.mockResolvedValue(mockCourses);

    render(<CourseListPage isAuthenticated={true} />);

    await waitFor(() => {
      expect(screen.getByText('Course Library')).toBeInTheDocument();
      expect(screen.getByText('Test Course 1')).toBeInTheDocument();
      expect(screen.getByText('Test Course 2')).toBeInTheDocument();
      expect(screen.getByText('Author: Author 1')).toBeInTheDocument();
      expect(screen.getByText('Language: en')).toBeInTheDocument();
      expect(screen.getByText('Duration: 2m 0s')).toBeInTheDocument();
    });
  });

  it('renders courses for public user', async () => {
    const { publicAPI } = require('../services/api');
    publicAPI.getCourses.mockResolvedValue(mockCourses);

    render(<CourseListPage isAuthenticated={false} />);

    await waitFor(() => {
      expect(screen.getByText('Course Library')).toBeInTheDocument();
      expect(screen.getByText('Test Course 1')).toBeInTheDocument();
      expect(screen.getByText('Test Course 2')).toBeInTheDocument();
    });
  });

  it('renders empty state when no courses', async () => {
    const { publicAPI } = require('../services/api');
    publicAPI.getCourses.mockResolvedValue([]);

    render(<CourseListPage isAuthenticated={false} />);

    await waitFor(() => {
      expect(screen.getByText('No courses found')).toBeInTheDocument();
      expect(screen.getByText('Public courses will appear here when available.')).toBeInTheDocument();
    });
  });

  it('renders error state on API failure', async () => {
    const { publicAPI } = require('../services/api');
    publicAPI.getCourses.mockRejectedValue(new Error('API Error'));

    render(<CourseListPage isAuthenticated={false} />);

    await waitFor(() => {
      expect(screen.getByText('Error: API Error')).toBeInTheDocument();
    });
  });

  it('formats dates correctly', async () => {
    const { publicAPI } = require('../services/api');
    publicAPI.getCourses.mockResolvedValue([mockCourses[0]]);

    render(<CourseListPage isAuthenticated={false} />);

    await waitFor(() => {
      expect(screen.getByText('Created: 1/1/2023')).toBeInTheDocument();
    });
  });
});