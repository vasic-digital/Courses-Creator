import React from 'react';
import styled from 'styled-components';
import { useNavigate } from 'react-router-dom';
import Helmet from 'react-helmet-async';

const PageContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 70vh;
  padding: ${props => props.theme.spacing.xl};
  text-align: center;
`;

const ErrorCode = styled.div`
  font-size: 6rem;
  font-weight: ${props => props.theme.typography.fontWeight.bold};
  color: ${props => props.theme.colors.primary.main};
  margin-bottom: ${props => props.theme.spacing.lg};
  line-height: 1;
`;

const Title = styled.h1`
  font-size: ${props => props.theme.typography.fontSize['2xl']};
  font-weight: ${props => props.theme.typography.fontWeight.semibold};
  color: ${props => props.theme.colors.text.primary};
  margin-bottom: ${props => props.theme.spacing.md};
`;

const Description = styled.p`
  font-size: ${props => props.theme.typography.fontSize.lg};
  color: ${props => props.theme.colors.text.secondary};
  margin-bottom: ${props => props.theme.spacing.xl};
  max-width: 500px;
  line-height: 1.6;
`;

const ButtonContainer = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: ${props => props.theme.spacing.md};
  justify-content: center;
`;

const Button = styled.button<{ variant?: 'primary' | 'secondary' }>`
  padding: ${props => props.theme.spacing.sm} ${props => props.theme.spacing.lg};
  border: none;
  border-radius: ${props => props.theme.borderRadius.md};
  font-size: ${props => props.theme.typography.fontSize.md};
  font-weight: ${props => props.theme.typography.fontWeight.medium};
  cursor: pointer;
  transition: background-color 0.2s ease;
  
  ${props => props.variant === 'primary' ? `
    background-color: ${props.theme.colors.primary.main};
    color: white;
    
    &:hover {
      background-color: ${props.theme.colors.primary.dark};
    }
  ` : `
    background-color: ${props.theme.colors.background.secondary};
    color: ${props.theme.colors.text.primary};
    
    &:hover {
      background-color: ${props.theme.colors.background.tertiary};
    }
  `}
`;

const NotFoundPage: React.FC = () => {
  const navigate = useNavigate();

  const handleGoHome = () => {
    navigate('/courses');
  };

  const handleGoBack = () => {
    window.history.back();
  };

  return (
    <>
      <Helmet>
        <title>Page Not Found - Course Player</title>
        <meta name="description" content="The page you're looking for doesn't exist" />
      </Helmet>

      <PageContainer>
        <ErrorCode>404</ErrorCode>
        <Title>Page Not Found</Title>
        <Description>
          The page you're looking for doesn't exist or has been moved. 
          Don't worry, we'll help you get back on track.
        </Description>
        
        <ButtonContainer>
          <Button variant="primary" onClick={handleGoHome}>
            Browse Courses
          </Button>
          <Button variant="secondary" onClick={handleGoBack}>
            Go Back
          </Button>
        </ButtonContainer>
      </PageContainer>
    </>
  );
};

export default NotFoundPage;