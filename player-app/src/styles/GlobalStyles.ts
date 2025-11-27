// Global styles
import { createGlobalStyle } from 'styled-components';

export const GlobalStyles = createGlobalStyle`
  * {
    box-sizing: border-box;
  }
  
  body {
    margin: 0;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
      'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    background-color: ${({ theme }) => theme.colors.background.primary};
    color: ${({ theme }) => theme.colors.text.primary};
    line-height: 1.5;
  }
  
  code {
    font-family: 'Fira Code', 'Monaco', 'Cascadia Code', 'Roboto Mono', monospace;
  }
  
  h1, h2, h3, h4, h5, h6 {
    margin: 0;
    font-weight: 600;
    color: ${({ theme }) => theme.colors.text.primary};
  }
  
  p {
    margin: 0;
  }
  
  a {
    color: ${({ theme }) => theme.colors.primary};
    text-decoration: none;
    transition: color 0.2s ease;
    
    &:hover {
      color: ${({ theme }) => theme.colors.primaryHover};
    }
  }
  
  button {
    font-family: inherit;
    cursor: pointer;
    border: none;
    background: none;
    padding: 0;
    margin: 0;
  }
  
  ul, ol {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  
  img {
    max-width: 100%;
    height: auto;
  }
  
  /* Scrollbar styles */
  ::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }
  
  ::-webkit-scrollbar-track {
    background: ${({ theme }) => theme.colors.background.secondary};
  }
  
  ::-webkit-scrollbar-thumb {
    background: ${({ theme }) => theme.colors.text.secondary};
    border-radius: 4px;
  }
  
  ::-webkit-scrollbar-thumb:hover {
    background: ${({ theme }) => theme.colors.text.primary};
  }
  
  /* Focus styles */
  :focus {
    outline: 2px solid ${({ theme }) => theme.colors.primary};
    outline-offset: 2px;
  }
  
  /* Selection styles */
  ::selection {
    background-color: ${({ theme }) => theme.colors.primary};
    color: white;
  }
  
  /* React Player styles */
  .react-player {
    position: relative;
    background-color: #000;
  }
  
  /* Code highlighting styles */
  .markdown-body {
    color: ${({ theme }) => theme.colors.text.primary};
    
    pre {
      background-color: ${({ theme }) => theme.colors.background.secondary};
      border-radius: 8px;
      padding: 16px;
      overflow-x: auto;
      margin: 16px 0;
    }
    
    code {
      background-color: ${({ theme }) => theme.colors.background.secondary};
      padding: 2px 6px;
      border-radius: 4px;
      font-size: 0.9em;
    }
    
    pre > code {
      background: none;
      padding: 0;
    }
    
    blockquote {
      border-left: 4px solid ${({ theme }) => theme.colors.primary};
      padding-left: 16px;
      margin: 16px 0;
      font-style: italic;
      color: ${({ theme }) => theme.colors.text.secondary};
    }
    
    table {
      border-collapse: collapse;
      width: 100%;
      margin: 16px 0;
    }
    
    th, td {
      border: 1px solid ${({ theme }) => theme.colors.border};
      padding: 8px 12px;
      text-align: left;
    }
    
    th {
      background-color: ${({ theme }) => theme.colors.background.secondary};
      font-weight: 600;
    }
  }
  
  /* Loading animation */
  @keyframes pulse {
    0% { opacity: 0.6; }
    50% { opacity: 1; }
    100% { opacity: 0.6; }
  }
  
  .pulse {
    animation: pulse 1.5s ease-in-out infinite;
  }
`;