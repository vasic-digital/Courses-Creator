import React, { useState } from 'react';
import { courseAPI } from '../services/api';
import './CourseCreation.css';

interface CourseCreationProps {
  onCourseCreated: () => void;
}

const CourseCreationPage: React.FC<CourseCreationProps> = ({ onCourseCreated }) => {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    author: '',
    language: 'en',
    tags: '',
    markdown: null as File | null,
    voice: 'alloy',
    outputDir: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, files } = e.target;
    if (files && files[0]) {
      setFormData(prev => ({ ...prev, [name]: files[0] }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      if (!formData.markdown) {
        throw new Error('Markdown file is required');
      }

      const options = {
        title: formData.title,
        description: formData.description,
        author: formData.author,
        language: formData.language,
        tags: formData.tags.split(',').map(tag => tag.trim()).filter(tag => tag),
        voice: formData.voice,
      };

      // For now, assume markdown is uploaded to a temp location
      // In real implementation, upload markdown first
      const markdownPath = `/tmp/${formData.markdown.name}`;
      const outputDir = formData.outputDir || `output/${formData.title.replace(/\s+/g, '_')}`;

      await courseAPI.generateCourse(markdownPath, outputDir, options);
      onCourseCreated();
    } catch (error: any) {
      setError(error.message || error.response?.data?.error || 'Failed to create course');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="course-creation">
      <h2>Create New Course</h2>
      <form onSubmit={handleSubmit} className="creation-form">
        <div className="form-group">
          <label htmlFor="title">Title *</label>
          <input
            type="text"
            id="title"
            name="title"
            value={formData.title}
            onChange={handleInputChange}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="description">Description *</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            rows={4}
            required
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="author">Author</label>
            <input
              type="text"
              id="author"
              name="author"
              value={formData.author}
              onChange={handleInputChange}
            />
          </div>

          <div className="form-group">
            <label htmlFor="language">Language</label>
            <select
              id="language"
              name="language"
              value={formData.language}
              onChange={handleInputChange}
            >
              <option value="en">English</option>
              <option value="es">Spanish</option>
              <option value="fr">French</option>
              <option value="de">German</option>
              <option value="it">Italian</option>
              <option value="pt">Portuguese</option>
              <option value="ru">Russian</option>
              <option value="ja">Japanese</option>
              <option value="ko">Korean</option>
              <option value="zh">Chinese</option>
            </select>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="tags">Tags (comma-separated)</label>
          <input
            type="text"
            id="tags"
            name="tags"
            value={formData.tags}
            onChange={handleInputChange}
            placeholder="e.g., programming, javascript, tutorial"
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="markdown">Markdown File *</label>
            <input
              type="file"
              id="markdown"
              name="markdown"
              accept=".md"
              onChange={handleFileChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="voice">Voice</label>
            <select
              id="voice"
              name="voice"
              value={formData.voice}
              onChange={handleInputChange}
            >
              <option value="alloy">Alloy</option>
              <option value="echo">Echo</option>
              <option value="fable">Fable</option>
              <option value="onyx">Onyx</option>
              <option value="nova">Nova</option>
              <option value="shimmer">Shimmer</option>
            </select>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="outputDir">Output Directory</label>
          <input
            type="text"
            id="outputDir"
            name="outputDir"
            value={formData.outputDir}
            onChange={handleInputChange}
            placeholder="output/course_name"
          />
        </div>

        {error && <div className="error-message">{error}</div>}

        <button type="submit" disabled={loading} className="submit-button">
          {loading ? 'Creating Course...' : 'Create Course'}
        </button>
      </form>
    </div>
  );
};

export default CourseCreationPage;