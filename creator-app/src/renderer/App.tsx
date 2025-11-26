import React, { useState } from 'react';
import axios from 'axios';

interface ProcessingOptions {
  voice?: string;
  backgroundMusic: boolean;
  languages: string[];
  quality: 'standard' | 'high';
}

const App: React.FC = () => {
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

      const response = await axios.post('http://localhost:8080/api/v1/courses/generate', {
        markdown_path: markdownFile,
        output_dir: outputDir,
        options: options
      });

      clearInterval(progressInterval);
      setProgress(100);

      setResult({
        type: 'success',
        message: `Course generated successfully! Output: ${response.data.output_path}`
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

  return (
    <div className="container">
      <h1>Course Creator</h1>

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
  );
};

export default App;