import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import Helmet from 'react-helmet-async';
import ReactPlayer from 'react-player';

import { apiService } from '@/services';
import { Course, Lesson } from '@/types';
import LoadingSpinner from '@/components/LoadingSpinner';

const PageContainer = styled.div`
  padding: ${props => props.theme.spacing.lg};
  max-width: 1200px;
  margin: 0 auto;
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

const CoursePlayerPage: React.FC = () => {
  const { courseId, lessonId } = useParams<{ courseId: string; lessonId?: string }>();
  const navigate = useNavigate();
  const [isPlaying, setIsPlaying] = useState(false);
  
  const {
    data: course,
    isLoading: courseLoading,
    error: courseError
  } = useQuery(
    ['course', courseId],
    () => courseId ? apiService.getCourse(courseId) : Promise.reject(new Error('Course ID is required')),
    {
      enabled: !!courseId,
      retry: 2,
      retryDelay: 1000,
    }
  );

  const getCurrentLesson = (): Lesson | undefined => {
    if (!course?.lessons) return undefined;
    
    if (lessonId) {
      return course.lessons.find(l => l.id === lessonId);
    }
    
    return course.lessons[0]; // Return first lesson if no specific lesson ID
  };

  const currentLesson = getCurrentLesson();

  const handleVideoEnd = () => {
    // Auto-play next lesson if available
    if (course && currentLesson) {
      const currentIndex = course.lessons.findIndex(l => l.id === currentLesson.id);
      if (currentIndex < course.lessons.length - 1) {
        const nextLesson = course.lessons[currentIndex + 1];
        navigate(`/courses/${courseId}/lessons/${nextLesson.id}`);
      }
    }
  };

  const handleLessonClick = (lessonId: string) => {
    navigate(`/courses/${courseId}/${lessonId}`);
  };

  const handleBack = () => {
    if (lessonId) {
      navigate(`/courses/${courseId}`);
    } else {
      navigate('/courses');
    }
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <>
      <Helmet>
        <title>{course?.title || 'Course'} - Course Player</title>
        <meta name="description" content={course?.description || 'Watch and learn from this video course'} />
      </Helmet>

      <PageContainer>
        <BackButton onClick={handleBack}>
          ← Back to {lessonId ? 'Course' : 'Courses'}
        </BackButton>

        {courseLoading && <LoadingSpinner message="Loading course..." />}

        {courseError && (
          <ErrorMessage>
            <p>Failed to load course. Please try again later.</p>
            <BackButton onClick={handleRefresh}>Try Again</BackButton>
          </ErrorMessage>
        )}

        {course && (
          <div>
            <div style={{ marginBottom: '20px' }}>
              <h1 style={{ margin: '0 0 10px 0' }}>{course.title}</h1>
              <p style={{ margin: '0', color: '#666' }}>{course.description}</p>
            </div>

            {course.lessons && course.lessons.length > 0 && (
              <div style={{ marginBottom: '20px' }}>
                <h2>Lessons</h2>
                <ul style={{ listStyle: 'none', padding: '0' }}>
                  {course.lessons.map((lesson) => (
                    <li 
                      key={lesson.id} 
                      style={{ 
                        padding: '10px', 
                        margin: '5px 0', 
                        backgroundColor: lesson.id === lessonId ? '#f0f0f0' : '#f9f9f9',
                        cursor: 'pointer',
                        border: lesson.id === lessonId ? '1px solid #ccc' : '1px solid #eee'
                      }}
                      onClick={() => handleLessonClick(lesson.id)}
                    >
                      <h3 style={{ margin: '0 0 5px 0' }}>{lesson.title}</h3>
                      <p style={{ margin: '0', fontSize: '14px', color: '#666' }}>
                        {lesson.duration ? `${Math.floor(lesson.duration / 60)}m ${lesson.duration % 60}s` : 'Unknown duration'}
                      </p>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {currentLesson && (
              <div>
                <h2>Now Playing: {currentLesson.title}</h2>
                <div style={{ 
                  width: '100%', 
                  maxWidth: '800px', 
                  backgroundColor: '#000', 
                  aspectRatio: '16/9',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                  fontSize: '18px',
                  marginBottom: '20px'
                }}>
                {currentLesson.videoUrl ? (
                  <ReactPlayer
                    url={currentLesson.videoUrl}
                    controls={true}
                    width="100%"
                    height="100%"
                    playing={isPlaying}
                    onPlay={() => setIsPlaying(true)}
                    onPause={() => setIsPlaying(false)}
                    onEnded={handleVideoEnd}
                    config={{
                      file: {
                        attributes: {
                          controlsList: 'nodownload',
                        }
                      }
                    }}
                  />
                ) : currentLesson.audioUrl ? (
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    width: '100%',
                    height: '100%',
                    padding: '20px'
                  }}>
                    <ReactPlayer
                      url={currentLesson.audioUrl}
                      controls={true}
                      width="100%"
                      height="50px"
                      playing={isPlaying}
                      onPlay={() => setIsPlaying(true)}
                      onPause={() => setIsPlaying(false)}
                      onEnded={handleVideoEnd}
                      config={{
                        file: {
                          attributes: {
                            controlsList: 'nodownload',
                          }
                        }
                      }}
                    />
                  </div>
                ) : (
                  <div style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: '100%',
                    color: 'white',
                    fontSize: '18px'
                  }}>
                    No media available for this lesson
                  </div>
                )}
                </div>
                
                {currentLesson.content && (
                  <div style={{ 
                    padding: '20px', 
                    backgroundColor: '#f9f9f9', 
                    borderRadius: '8px' 
                  }}>
                    <h3>Lesson Content</h3>
                    <div dangerouslySetInnerHTML={{ __html: currentLesson.content }} />
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </PageContainer>
    </>
  );
};

export default CoursePlayerPage;