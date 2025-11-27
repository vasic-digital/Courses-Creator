import React, { useState, useEffect } from 'react';
import { publicAPI, courseAPI } from '../services/api';
import './CourseList.css';

interface Course {
  id: string;
  title: string;
  description: string;
  created_at: string;
  updated_at: string;
  metadata?: {
    author?: string;
    language?: string;
    total_duration?: number;
  };
}

interface CourseListProps {
  isAuthenticated: boolean;
}

const CourseListPage: React.FC<CourseListProps> = ({ isAuthenticated }) => {
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const data = await (isAuthenticated ? courseAPI.getCourses() : publicAPI.getCourses());
        setCourses(data);
      } catch (error: any) {
        setError(error.response?.data?.error || 'Failed to fetch courses');
      } finally {
        setLoading(false);
      }
    };

    fetchCourses();
  }, [isAuthenticated]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  if (loading) {
    return <div className="course-list-loading">Loading courses...</div>;
  }

  if (error) {
    return <div className="course-list-error">Error: {error}</div>;
  }

  if (courses.length === 0) {
    return (
      <div className="course-list-empty">
        <h3>No courses found</h3>
        <p>
          {isAuthenticated 
            ? "Create your first course to get started." 
            ? "Public courses will appear here when available."
            : "Public courses will appear here when available."
          }
        </p>
      </div>
    );
  }

  return (
    <div className="course-list">
      <h2>Course Library</h2>
      <div className="course-grid">
        {courses.map((course) => (
          <div key={course.id} className="course-card">
            <div className="course-header">
              <h3>{course.title}</h3>
            </div>
            <div className="course-body">
              <p className="course-description">{course.description}</p>
              <div className="course-metadata">
                {course.metadata?.author && (
                  <span className="metadata-item">
                    <strong>Author:</strong> {course.metadata.author}
                  </span>
                )}
                {course.metadata?.language && (
                  <span className="metadata-item">
                    <strong>Language:</strong> {course.metadata.language}
                  </span>
                )}
                {course.metadata?.total_duration && (
                  <span className="metadata-item">
                    <strong>Duration:</strong> {Math.floor(course.metadata.total_duration / 60)}m {course.metadata.total_duration % 60}s
                  </span>
                )}
              </div>
              <div className="course-dates">
                <span className="date-item">
                  <strong>Created:</strong> {formatDate(course.created_at)}
                </span>
                {course.updated_at !== course.created_at && (
                  <span className="date-item">
                    <strong>Updated:</strong> {formatDate(course.updated_at)}
                  </span>
                )}
              </div>
            </div>
            <div className="course-footer">
              <button className="play-button">
                Play Course
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CourseListPage;