import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Helmet from 'react-helmet-async';

import { apiService } from '@/services';
import { Course, Lesson } from '@/types';
import LoadingSpinner from '@/components/LoadingSpinner';

const PageContainer = styled.div`
  padding: ${props => props.theme.spacing.lg};
  max-width: 1200px;
  margin: 0 auto;
`;

const CourseHeader = styled.div`
  margin-bottom: ${props => props.theme.spacing.xl};
  text-align: center;
`;

const CourseTitle = styled.h1`
  font-size: ${props => props.theme.typography.fontSize['3xl']};
  font-weight: ${props => props.theme.typography.fontWeight.bold};
  color: ${props => props.theme.colors.text.primary};
  margin-bottom: ${props => props.theme.spacing.md};
`;

const CourseDescription = styled.p`
  font-size: ${props => props.theme.typography.fontSize.lg};
  color: ${props => props.theme.colors.text.secondary};
  max-width: 800px;
  margin: 0 auto ${props => props.theme.spacing.xl};
  line-height: 1.6;
`;

const ErrorMessage = styled.div`
  background-color: ${props => props.theme.colors.error.background};
  color: ${props => props.theme.colors.error.text};
  padding: ${props => props.theme.spacing.lg};
  border-radius: ${props => props.theme.borderRadius.md};
  margin: ${props => props.theme.spacing.lg} 0;
  text-align: center;
`;

const BackButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.colors.primary.main};
  font-size: ${props => props.theme.typography.fontSize.md};
  cursor: pointer;
  display: flex;
  align-items: center;
  margin-bottom: ${props => props.theme.spacing.lg};
  
  &:hover {
    text-decoration: underline;
  }
`;

const CourseListPage: React.FC = () => {
  const navigate = useNavigate();
  
  const {
    data: coursesResponse,
    isLoading,
    error,
    refetch
  } = useQuery(
    'courses',
    () => apiService.getCourses(),
    {
      retry: 2,
      retryDelay: 1000,
    }
  );

  const courses = coursesResponse?.items || [];

  const handleCourseClick = (courseId: string) => {
    navigate(`/courses/${courseId}`);
  };

  const handleRefresh = () => {
    refetch();
  };

  return (
    <>
      <Helmet>
        <title>Course Library - Course Player</title>
        <meta name="description" content="Browse and watch video courses on various topics" />
      </Helmet>

      <PageContainer>
        <CourseHeader>
          <CourseTitle>Course Library</CourseTitle>
          <CourseDescription>
            Browse our collection of video courses and start learning today
          </CourseDescription>
        </CourseHeader>

        {isLoading && <LoadingSpinner message="Loading courses..." />}

        {error && (
          <ErrorMessage>
            <p>Failed to load courses. Please try again later.</p>
            <BackButton onClick={handleRefresh}>Try Again</BackButton>
          </ErrorMessage>
        )}

        {courses && (
          <div>
            {courses.length === 0 ? (
              <ErrorMessage>
                <p>No courses available at the moment.</p>
                <p>Check back later for new content!</p>
              </ErrorMessage>
            ) : (
              <div>
                {/* Course grid will be implemented here */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '20px' }}>
                  {courses.map((course: Course) => (
                    <div 
                      key={course.id}
                      style={{
                        border: '1px solid #ccc',
                        borderRadius: '8px',
                        padding: '16px',
                        cursor: 'pointer',
                        backgroundColor: '#f9f9f9'
                      }}
                      onClick={() => handleCourseClick(course.id)}
                    >
                      <h3 style={{ margin: '0 0 8px 0' }}>{course.title}</h3>
                      <p style={{ margin: '0 0 8px 0', color: '#666' }}>
                        {course.description}
                      </p>
                      <p style={{ margin: '0', fontSize: '14px', color: '#888' }}>
                        {course.lessons?.length || 0} lessons
                      </p>
                      <p style={{ margin: '4px 0 0 0', fontSize: '12px', color: '#aaa' }}>
                        {course.metadata?.estimatedHours ? `${course.metadata.estimatedHours}h` : 'Unknown duration'}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </PageContainer>
    </>
  );
};

export default CourseListPage;