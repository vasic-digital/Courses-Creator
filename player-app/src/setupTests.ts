// jest-dom adds custom jest matchers for asserting on DOM nodes.
import '@testing-library/jest-dom';

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock ReactPlayer
jest.mock('react-player', () => {
  return jest.fn(({ url, playing }) => (
    <div data-testid="react-player" data-url={url} data-playing={playing}>
      Video Player
    </div>
  ));
});