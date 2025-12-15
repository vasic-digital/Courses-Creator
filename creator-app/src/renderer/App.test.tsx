import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import App from "./App";

// Mock the API services
jest.mock("./services/api", () => ({
  authAPI: {
    getCurrentUser: jest.fn(),
    logout: jest.fn(),
  },
  courseAPI: {
    generateCourse: jest.fn(),
  },
}));

// Mock Electron API
const mockElectronAPI = {
  selectMarkdownFile: jest.fn(),
  selectOutputDirectory: jest.fn(),
  readFile: jest.fn(),
};

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  removeItem: jest.fn(),
  setItem: jest.fn(),
  clear: jest.fn(),
};

Object.defineProperty(window, "localStorage", {
  value: mockLocalStorage,
});

Object.defineProperty(window, "electronAPI", {
  value: mockElectronAPI,
});

describe("App Component", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockLocalStorage.getItem.mockReturnValue(null);
  });

  it("renders login page when not authenticated", () => {
    render(<App />);

    // Should show login page
    expect(screen.getByRole("heading", { name: /Login/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/Email:/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password:/i)).toBeInTheDocument();
  });

  it("renders course list when authenticated", async () => {
    // Mock successful authentication
    const { authAPI } = require("./services/api");
    authAPI.getCurrentUser.mockResolvedValue({});
    mockLocalStorage.getItem.mockReturnValue("fake-token");

    render(<App />);

    // Wait for authentication check
    await waitFor(() => {
      expect(screen.getByText(/My Courses/i)).toBeInTheDocument();
      expect(screen.getByText(/Create Course/i)).toBeInTheDocument();
    });
  });

  it("handles file selection", async () => {
    // Mock authenticated state
    const { authAPI } = require("./services/api");
    authAPI.getCurrentUser.mockResolvedValue({});
    mockLocalStorage.getItem.mockReturnValue("fake-token");

    // Mock file selection
    mockElectronAPI.selectMarkdownFile.mockResolvedValue("/path/to/file.md");
    mockElectronAPI.readFile.mockResolvedValue("# Test Content");

    render(<App />);

    // Switch to creator page
    await waitFor(() => {
      fireEvent.click(screen.getByText(/Create Course/i));
    });

    // Click file selection button
    fireEvent.click(screen.getByText(/Select Markdown File/i));

    await waitFor(() => {
      expect(mockElectronAPI.selectMarkdownFile).toHaveBeenCalled();
    });
  });

  it("handles logout", async () => {
    // Mock authenticated state
    const { authAPI } = require("./services/api");
    authAPI.getCurrentUser.mockResolvedValue({});
    mockLocalStorage.getItem.mockReturnValue("fake-token");

    render(<App />);

    // Wait for authentication
    await waitFor(() => {
      expect(screen.getByText(/Logout/i)).toBeInTheDocument();
    });

    // Click logout
    fireEvent.click(screen.getByText(/Logout/i));

    expect(authAPI.logout).toHaveBeenCalled();
  });

  it("disables generate button when no files are selected", async () => {
    // Mock authenticated state
    const { authAPI } = require("./services/api");
    authAPI.getCurrentUser.mockResolvedValue({});
    mockLocalStorage.getItem.mockReturnValue("fake-token");

    render(<App />);

    // Switch to creator page
    await waitFor(() => {
      fireEvent.click(screen.getByText(/Create Course/i));
    });

    // Generate button should be disabled when no files are selected
    const generateButton = screen.getByText(/Generate Course/i);
    expect(generateButton).toBeDisabled();
  });
});
