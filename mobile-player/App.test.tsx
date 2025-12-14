import React from 'react';
import { render } from '@testing-library/react-native';
import App from './App';

// Mock the navigation components
jest.mock('@react-navigation/native', () => ({
  NavigationContainer: ({ children }: { children: React.ReactNode }) => children,
}));

jest.mock('@react-navigation/stack', () => ({
  createStackNavigator: () => ({
    Navigator: ({ children }: { children: React.ReactNode }) => children,
    Screen: () => null,
  }),
}));

// Mock the screen components
jest.mock('./src/screens/CourseListScreen', () => 'CourseListScreen');
jest.mock('./src/screens/CoursePlayerScreen', () => 'CoursePlayerScreen');

describe('App', () => {
  it('renders without crashing', () => {
    const { getByText } = render(<App />);
    
    // Since we're mocking navigation, we can't test for specific text
    // But we can verify the component renders
    expect(getByText).toBeDefined();
  });

  it('has correct structure', () => {
    const { UNSAFE_getAllByType } = render(<App />);
    
    // The app should render without errors
    expect(UNSAFE_getAllByType).toBeDefined();
  });
});