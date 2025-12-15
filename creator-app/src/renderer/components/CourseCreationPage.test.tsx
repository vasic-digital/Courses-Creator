import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import CourseCreationPage from './CourseCreationPage';

// Mock the API
jest.mock('../services/api', () => ({
  courseAPI: {
    generateCourse: jest.fn(),
  },
}));

const mockOnCourseCreated = jest.fn();

describe('CourseCreationPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders the course creation form', () => {
    render(<CourseCreationPage onCourseCreated={mockOnCourseCreated} />);

    expect(screen.getByText('Create New Course')).toBeInTheDocument();
    expect(screen.getByLabelText('Title *')).toBeInTheDocument();
    expect(screen.getByLabelText('Description *')).toBeInTheDocument();
    expect(screen.getByLabelText('Author')).toBeInTheDocument();
    expect(screen.getByLabelText('Language')).toBeInTheDocument();
    expect(screen.getByLabelText('Tags (comma-separated)')).toBeInTheDocument();
    expect(screen.getByLabelText('Markdown File *')).toBeInTheDocument();
    expect(screen.getByLabelText('Voice')).toBeInTheDocument();
    expect(screen.getByLabelText('Output Directory')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Create Course' })).toBeInTheDocument();
  });

  it('submits the form with valid data', async () => {
    const { courseAPI } = require('../services/api');
    courseAPI.generateCourse.mockResolvedValue({});

    render(<CourseCreationPage onCourseCreated={mockOnCourseCreated} />);

    // Fill out the form
    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'Test Course' } });
    fireEvent.change(screen.getByLabelText('Description *'), { target: { value: 'Test Description' } });
    fireEvent.change(screen.getByLabelText('Author'), { target: { value: 'Test Author' } });
    fireEvent.change(screen.getByLabelText('Tags (comma-separated)'), { target: { value: 'test, course' } });

    // Mock file input
    const file = new File(['test content'], 'test.md', { type: 'text/markdown' });
    const fileInput = screen.getByLabelText('Markdown File *');
    fireEvent.change(fileInput, { target: { files: [file] } });

    fireEvent.change(screen.getByLabelText('Output Directory'), { target: { value: 'output/test' } });

    // Submit the form
    fireEvent.submit(screen.getByRole('button', { name: 'Create Course' }).closest('form')!);

    await waitFor(() => {
      expect(courseAPI.generateCourse).toHaveBeenCalledWith(
        '/tmp/test.md', // This would be the actual path in real implementation
        'output/test',
        {
          title: 'Test Course',
          description: 'Test Description',
          author: 'Test Author',
          language: 'en',
          tags: ['test', 'course'],
          voice: 'alloy',
        }
      );
      expect(mockOnCourseCreated).toHaveBeenCalled();
    });
  });

  it('shows error when markdown file is not provided', async () => {
    render(<CourseCreationPage onCourseCreated={mockOnCourseCreated} />);

    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'Test Course' } });
    fireEvent.change(screen.getByLabelText('Description *'), { target: { value: 'Test Description' } });

    // Submit without file
    fireEvent.submit(screen.getByRole('button', { name: 'Create Course' }).closest('form')!);

    await waitFor(() => {
      expect(screen.getByText('Markdown file is required')).toBeInTheDocument();
    });
  });

  it('displays loading state during submission', async () => {
    const { courseAPI } = require('../services/api');
    courseAPI.generateCourse.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    render(<CourseCreationPage onCourseCreated={mockOnCourseCreated} />);

    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'Test Course' } });
    fireEvent.change(screen.getByLabelText('Description *'), { target: { value: 'Test Description' } });

    const file = new File(['test'], 'test.md');
    fireEvent.change(screen.getByLabelText('Markdown File *'), { target: { files: [file] } });

    fireEvent.submit(screen.getByRole('button', { name: 'Create Course' }).closest('form')!);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Creating Course...' })).toBeInTheDocument();
      expect(screen.getByRole('button')).toBeDisabled();
    });
  });

  it('handles API errors', async () => {
    const { courseAPI } = require('../services/api');
    courseAPI.generateCourse.mockRejectedValue(new Error('API Error'));

    render(<CourseCreationPage onCourseCreated={mockOnCourseCreated} />);

    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'Test Course' } });
    fireEvent.change(screen.getByLabelText('Description *'), { target: { value: 'Test Description' } });

    const file = new File(['test'], 'test.md');
    fireEvent.change(screen.getByLabelText('Markdown File *'), { target: { files: [file] } });

    fireEvent.submit(screen.getByRole('button', { name: 'Create Course' }).closest('form')!);

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });
});