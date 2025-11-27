import React from 'react';
import styled from 'styled-components';

const SpinnerContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: ${props => props.theme.spacing.lg};
  width: 100%;
  height: 100%;
`;

const Spinner = styled.div`
  width: 40px;
  height: 40px;
  border: 4px solid ${props => props.theme.colors.border.light};
  border-top: 4px solid ${props => props.theme.colors.primary.main};
  border-radius: 50%;
  animation: spin 1s linear infinite;
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;

const LoadingText = styled.p`
  margin-top: ${props => props.theme.spacing.md};
  color: ${props => props.theme.colors.text.secondary};
  font-size: ${props => props.theme.typography.fontSize.md};
`;

interface LoadingSpinnerProps {
  message?: string;
  size?: 'small' | 'medium' | 'large';
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  message = 'Loading...', 
  size = 'medium' 
}) => {
  const getSize = () => {
    switch (size) {
      case 'small': return '24px';
      case 'large': return '56px';
      default: return '40px';
    }
  };

  return (
    <SpinnerContainer>
      <div>
        <Spinner size={getSize()} />
        {message && <LoadingText>{message}</LoadingText>}
      </div>
    </SpinnerContainer>
  );
};

export default LoadingSpinner;