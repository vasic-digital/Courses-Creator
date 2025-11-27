import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import styled from 'styled-components';
import { Helmet } from 'react-helmet-async';

// Pages
import CourseListPage from '@/pages/CourseListPage';
import CoursePlayerPage from '@/pages/CoursePlayerPage';
import NotFoundPage from '@/pages/NotFoundPage';

// Components
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import LoadingSpinner from '@/components/LoadingSpinner';

const AppContainer = styled.div`
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: ${props => props.theme.colors.background.primary};
`;

const MainContent = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
`;

const App: React.FC = () => {
  return (
    <AppContainer>
      <Helmet>
        <title>Course Player - Learn from Video Courses</title>
        <meta name="description" content="Watch and learn from high-quality video courses created with Course Creator" />
      </Helmet>
      
      <Header />
      
      <MainContent>
        <Routes>
          <Route path="/" element={<Navigate to="/courses" replace />} />
          <Route path="/courses" element={<CourseListPage />} />
          <Route path="/courses/:courseId" element={<CoursePlayerPage />} />
          <Route path="/courses/:courseId/:lessonId" element={<CoursePlayerPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </MainContent>
      
      <Footer />
    </AppContainer>
  );
};

export default App;