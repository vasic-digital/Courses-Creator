import React, { useState, useEffect } from 'react';
import { authAPI } from './services/api';
import LoginPage from './components/LoginPage';
import CourseListPage from './components/CourseListPage';
import './App.css';

interface ProcessingOptions {
  voice?: string;
  backgroundMusic: boolean;
  languages: string[];
  quality: 'standard' | 'high';
}

const App: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [currentPage, setCurrentPage] = useState<'courses' | 'creator'>('courses');
  const [markdownFile, setMarkdownFile] = useState<string>('');
  const [outputDir, setOutputDir] = useState<string>('');
  const [markdownContent, setMarkdownContent] = useState<string>('');
  const [options, setOptions] = useState<ProcessingOptions>({
    backgroundMusic: false,
    languages: ['en'],
    quality: 'standard'
  });
  const [isProcessing, setIsProcessing] = useState<boolean>(false);
  const [progress, setProgress] = useState<number>(0);
  const [result, setResult] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      // Verify token is valid by fetching current user
      authAPI.getCurrentUser()
        .then(() => setIsAuthenticated(true))
        .catch(() => {
          localStorage.removeItem('auth_token');
          setIsAuthenticated(false);
        });
    }
  }, []);

  const handleAuthSuccess = () => {
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    authAPI.logout();
    setIsAuthenticated(false);
    setCurrentPage('courses');
  };

  const selectMarkdownFile = async () => {
    try {
      const filePath = await (window as any).electronAPI.selectMarkdownFile();
      if (filePath) {
        setMarkdownFile(filePath);
        const content = await (window as any).electronAPI.readFile(filePath);
        setMarkdownContent(content);
      }
    } catch (error) {
      console.error('Error selecting file:', error);
    }
  };

  const selectOutputDirectory = async () => {
    try {
      const dirPath = await (window as any).electronAPI.selectOutputDirectory();
      if (dirPath) {
        setOutputDir(dirPath);
      }
    } catch (error) {
      console.error('Error selecting directory:', error);
    }
  };

  const generateCourse = async () => {
    if (!markdownFile || !outputDir) {
      setResult({ type: 'error', message: 'Please select both markdown file and output directory.' });
      return;
    }

    setIsProcessing(true);
    setProgress(0);
    setResult(null);

    try {
      // Simulate progress
      const progressInterval = setInterval(() => {
        setProgress(prev => Math.min(prev + 10, 90));
      }, 500);

      // Import dynamically to avoid issues when not authenticated
      const { courseAPI } = await import('./services/api');
      const response = await courseAPI.generateCourse(markdownFile, outputDir, options);

      clearInterval(progressInterval);
      setProgress(100);

      setResult({
        type: 'success',
        message: `Course generated successfully! Job ID: ${response.data.job_id || 'unknown'}`
      });
    } catch (error: any) {
      setResult({
        type: 'error',
        message: error.response?.data?.error || 'Failed to generate course.'
      });
    } finally {
      setIsProcessing(false);
    }
  };

  if (!isAuthenticated) {
    return <LoginPage onAuthSuccess={handleAuthSuccess} />;
  }

  return (
    <div className="app-container">
      <nav className="navbar">
        <div className="nav-brand">Course Creator</div>
        <div className="nav-links">
          <button 
            className={currentPage === 'courses' ? 'nav-link active' : 'nav-link'}
            onClick={() => setCurrentPage('courses')}
          >
            My Courses
          </button>
          <button 
            className={currentPage === 'creator' ? 'nav-link active' : 'nav-link'}
            onClick={() => setCurrentPage('creator')}
          >
            Create Course
          </button>
          <button className="nav-link logout" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </nav>

      <main className="main-content">
        {currentPage === 'courses' ? (
          <CourseListPage isAuthenticated={isAuthenticated} />
        ) : (
          <div className="creator-container">
            <h1>Create New Course</h1>

            <div className="form-group">
              <label>Markdown File:</label>
              <button onClick={selectMarkdownFile} disabled={isProcessing}>
                {markdownFile ? 'Change File' : 'Select Markdown File'}
              </button>
              {markdownFile && (
                <div className="file-info">
                  Selected: {markdownFile.split('/').pop()}
                </div>
              )}
            </div>

            <div className="form-group">
              <label>Output Directory:</label>
              <button onClick={selectOutputDirectory} disabled={isProcessing}>
                {outputDir ? 'Change Directory' : 'Select Output Directory'}
              </button>
              {outputDir && (
                <div className="file-info">
                  Selected: {outputDir}
                </div>
              )}
            </div>

            <div className="form-group">
              <label>Markdown Preview:</label>
              <textarea
                value={markdownContent}
                onChange={(e) => setMarkdownContent(e.target.value)}
                placeholder="Your markdown content will appear here..."
                disabled={isProcessing}
              />
            </div>

            <div className="form-group">
              <label>Voice:</label>
              <select
                value={options.voice || ''}
                onChange={(e) => setOptions({ ...options, voice: e.target.value || undefined })}
                disabled={isProcessing}
              >
                <option value="">Default</option>
                <option value="bark">Bark</option>
                <option value="speecht5">SpeechT5</option>
              </select>
            </div>

            <div className="form-group">
              <label>
                <input
                  type="checkbox"
                  checked={options.backgroundMusic}
                  onChange={(e) => setOptions({ ...options, backgroundMusic: e.target.checked })}
                  disabled={isProcessing}
                />
                Background Music
              </label>
            </div>

            <div className="form-group">
              <label>Quality:</label>
              <select
                value={options.quality}
                onChange={(e) => setOptions({ ...options, quality: e.target.value as 'standard' | 'high' })}
                disabled={isProcessing}
              >
                <option value="standard">Standard</option>
                <option value="high">High</option>
              </select>
            </div>

            <button onClick={generateCourse} disabled={isProcessing || !markdownFile || !outputDir}>
              {isProcessing ? 'Generating Course...' : 'Generate Course'}
            </button>

            {isProcessing && (
              <div className="progress">
                <div className="progress-bar">
                  <div className="progress-fill" style={{ width: `${progress}%` }}></div>
                </div>
                <p>Processing... {progress}%</p>
              </div>
            )}

            {result && (
              <div className={`result ${result.type}`}>
                <strong>{result.type === 'success' ? 'Success!' : 'Error:'}</strong> {result.message}
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
};

export default App;