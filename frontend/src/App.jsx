import React, { useState, useEffect, useRef } from 'react';
import { 
  Upload, 
  Search, 
  Settings as SettingsIcon, 
  FileText, 
  Server, 
  CheckCircle, 
  AlertCircle, 
  Database,
  ArrowRight,
  BookOpen,
  HelpCircle,
  Sun,
  Moon
} from 'lucide-react';
import './App.css';

const API_BASE = `http://${window.location.hostname}:8080`;

function App() {
  // Theme State
  const [theme, setTheme] = useState(() => {
    const saved = localStorage.getItem('theme');
    return saved || 'light';
  });

  // Apply theme to document root
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light');
  };

  // Configuration State
  const [chunkSize, setChunkSize] = useState(1000);
  const [chunkOverlap, setChunkOverlap] = useState(150);

  // Status Indicators
  const [backendStatus, setBackendStatus] = useState('checking'); // 'checking', 'online', 'offline'
  
  // Upload States
  const [isDragging, setIsDragging] = useState(false);
  const [uploadStatus, setUploadStatus] = useState('idle'); // 'idle', 'uploading', 'processing', 'success', 'error'
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploadErrorMsg, setUploadErrorMsg] = useState('');
  const [processedFile, setProcessedFile] = useState(null);
  
  // Knowledge Base State (Local History for display)
  const [indexedFiles, setIndexedFiles] = useState([]);

  // Search State
  const [query, setQuery] = useState('');
  const [limit, setLimit] = useState(5);
  const [searchResults, setSearchResults] = useState(null);
  const [isSearching, setIsSearching] = useState(false);

  const fileInputRef = useRef(null);

  // Health check on mount and interval
  useEffect(() => {
    const checkHealth = async () => {
      try {
        const response = await fetch(`${API_BASE}/api/health`);
        if (response.ok) {
          setBackendStatus('online');
        } else {
          setBackendStatus('offline');
        }
      } catch (err) {
        setBackendStatus('offline');
      }
    };

    checkHealth();
    const interval = setInterval(checkHealth, 5000);
    return () => clearInterval(interval);
  }, []);

  // Drag and drop handlers
  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      processUpload(files[0]);
    }
  };

  const handleFileSelect = (e) => {
    const files = e.target.files;
    if (files.length > 0) {
      processUpload(files[0]);
    }
  };

  // Perform backend PDF Upload and Ingestion
  const processUpload = async (file) => {
    if (file.type !== 'application/pdf') {
      setUploadStatus('error');
      setUploadErrorMsg('Only PDF documents are supported.');
      return;
    }

    setUploadStatus('uploading');
    setUploadProgress(15);
    setUploadErrorMsg('');

    const formData = new FormData();
    formData.append('file', file);
    formData.append('chunk_size', chunkSize.toString());
    formData.append('chunk_overlap', chunkOverlap.toString());

    try {
      // Simulate progress up to 60% before server handles ingestion
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev < 60) return prev + 10;
          clearInterval(progressInterval);
          return 60;
        });
      }, 300);

      const response = await fetch(`${API_BASE}/api/upload`, {
        method: 'POST',
        body: formData,
      });

      clearInterval(progressInterval);

      if (!response.ok) {
        const errText = await response.text();
        throw new Error(errText || 'Failed to process document');
      }

      const result = await response.json();
      setUploadProgress(100);
      setUploadStatus('success');
      setProcessedFile({
        name: result.filename,
        chunks: result.chunks_count
      });

      // Append to local state list
      setIndexedFiles(prev => [
        { name: result.filename, chunks: result.chunks_count, id: Date.now() },
        ...prev
      ]);
    } catch (err) {
      setUploadStatus('error');
      setUploadErrorMsg(err.message || 'An error occurred during upload.');
    }
  };

  // Search logic
  const handleSearch = async (e) => {
    if (e) e.preventDefault();
    if (!query.trim()) return;

    setIsSearching(true);
    try {
      const response = await fetch(`${API_BASE}/api/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: query,
          limit: parseInt(limit, 10),
        }),
      });

      if (!response.ok) {
        throw new Error('Search request failed');
      }

      const data = await response.json();
      setSearchResults(data.results || []);
    } catch (err) {
      console.error(err);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <div className="app-container">
      {/* Header */}
      <header className="app-header glass-panel">
        <div className="brand-section">
          <Database className="brand-logo" size={28} />
          <h1>
            ARIA <span className="gradient-text">Knowledge Hub</span>
          </h1>
        </div>
        
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <button 
            onClick={toggleTheme} 
            className="theme-toggle-btn"
            title={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
            aria-label="Toggle theme"
          >
            {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
          </button>
          
          <div className="status-badge">
            <Server size={14} />
            <span>Backend status:</span>
            <span className={`status-dot ${backendStatus === 'online' ? 'online' : backendStatus === 'offline' ? 'error' : ''}`} />
            <span style={{ textTransform: 'capitalize' }}>{backendStatus}</span>
          </div>
        </div>
      </header>

      {/* Main Area */}
      <main className="app-main">
        
        {/* Left Sidebar for Upload & Settings */}
        <section className="app-sidebar">
          
          {/* Settings Section */}
          <div className="sidebar-section glass-panel">
            <h2>
              <SettingsIcon size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
              Ingestion Configs
            </h2>
            <div className="form-group">
              <label>Chunk Size (characters)</label>
              <input 
                type="number" 
                value={chunkSize} 
                onChange={(e) => setChunkSize(parseInt(e.target.value, 10) || 0)} 
                min="100"
                max="5000"
              />
            </div>
            <div className="form-group">
              <label>Chunk Overlap (characters)</label>
              <input 
                type="number" 
                value={chunkOverlap} 
                onChange={(e) => setChunkOverlap(parseInt(e.target.value, 10) || 0)}
                min="0"
                max="1000"
              />
            </div>
          </div>

          {/* Upload Section */}
          <div className="sidebar-section glass-panel" style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
            <h2>
              <Upload size={16} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
              Upload PDF
            </h2>
            
            <div 
              className={`upload-zone ${isDragging ? 'dragging' : ''}`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <input 
                type="file" 
                ref={fileInputRef} 
                onChange={handleFileSelect} 
                accept="application/pdf" 
                style={{ display: 'none' }} 
              />
              <FileText className="upload-icon" size={40} />
              <p className="upload-text">Drag & drop your PDF here</p>
              <p className="upload-subtext">or click to browse local files</p>
            </div>

            {/* Ingestion Progress / Success Status */}
            {uploadStatus !== 'idle' && (
              <div className="upload-progress-container" style={{ marginTop: '1rem' }}>
                <div className="progress-header">
                  <span>
                    {uploadStatus === 'uploading' && 'Uploading...'}
                    {uploadStatus === 'processing' && 'Extracting & Embedding...'}
                    {uploadStatus === 'success' && 'Ingested Successfully!'}
                    {uploadStatus === 'error' && 'Ingestion Failed'}
                  </span>
                  <span>{uploadStatus !== 'error' && `${uploadProgress}%`}</span>
                </div>
                
                {uploadStatus !== 'success' && uploadStatus !== 'error' && (
                  <div className="progress-bar-bg">
                    <div 
                      className={`progress-bar-fill ${uploadStatus === 'processing' ? 'pulse' : ''}`}
                      style={{ width: `${uploadProgress}%` }}
                    />
                  </div>
                )}

                {uploadStatus === 'success' && processedFile && (
                  <div style={{ color: 'var(--secondary)', fontSize: '0.8rem', display: 'flex', alignItems: 'center', gap: '4px', marginTop: '4px' }}>
                    <CheckCircle size={14} />
                    <span>Processed {processedFile.name} ({processedFile.chunks} chunks)</span>
                  </div>
                )}

                {uploadStatus === 'error' && (
                  <div style={{ color: '#ef4444', fontSize: '0.8rem', display: 'flex', alignItems: 'flex-start', gap: '4px', marginTop: '4px' }}>
                    <AlertCircle size={14} style={{ flexShrink: 0, marginTop: '2px' }} />
                    <span>{uploadErrorMsg}</span>
                  </div>
                )}
              </div>
            )}

            {/* Document Registry List */}
            <div style={{ marginTop: '1.5rem', flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
              <h3 style={{ fontSize: '0.85rem', color: 'var(--text-gray)', fontWeight: '600', marginBottom: '0.5rem' }}>
                Indexed Files
              </h3>
              
              <div className="doc-list">
                {indexedFiles.length === 0 ? (
                  <div style={{ textAlign: 'center', padding: '1.5rem 0', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                    No documents indexed yet
                  </div>
                ) : (
                  indexedFiles.map(file => (
                    <div key={file.id} className="doc-item">
                      <BookOpen size={16} style={{ color: 'var(--primary)' }} />
                      <div className="doc-details">
                        <div className="doc-name">{file.name}</div>
                        <div className="doc-meta">{file.chunks} vector chunks</div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

          </div>
        </section>

        {/* Right Search panel */}
        <section className="glass-panel search-container">
          <div className="search-bar-section">
            <form onSubmit={handleSearch} className="search-input-wrapper">
              <Search className="search-icon" size={20} />
              <input 
                type="text" 
                placeholder="Ask a question or search your knowledge base..."
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                disabled={backendStatus !== 'online'}
              />
              <button 
                type="submit" 
                style={{
                  position: 'absolute',
                  right: '8px',
                  background: 'var(--primary)',
                  border: 'none',
                  borderRadius: '8px',
                  color: 'white',
                  padding: '0.6rem 1rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px',
                  cursor: 'pointer',
                  fontWeight: '600',
                  fontSize: '0.85rem'
                }}
                disabled={isSearching || !query.trim() || backendStatus !== 'online'}
              >
                Search <ArrowRight size={14} />
              </button>
            </form>

            <div className="search-options">
              <div className="slider-group">
                <label>Maximum results to return:</label>
                <input 
                  type="range" 
                  min="1" 
                  max="10" 
                  value={limit} 
                  onChange={(e) => setLimit(parseInt(e.target.value, 10))} 
                />
                <span style={{ fontSize: '0.85rem', fontWeight: '600' }}>{limit}</span>
              </div>
              
              <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                Powered by pgvector & nomic-embed-text
              </div>
            </div>
          </div>

          {/* Search results list */}
          <div className="results-viewport">
            {isSearching ? (
              <div className="results-status-info">
                <div style={{
                  border: '4px solid var(--border-subtle)',
                  borderTop: '4px solid var(--primary)',
                  borderRadius: '50%',
                  width: '40px',
                  height: '40px',
                  animation: 'progressGlow 1s infinite linear'
                }} />
                <p>Retrieving matching document chunks...</p>
              </div>
            ) : searchResults === null ? (
              <div className="results-status-info">
                <HelpCircle size={48} style={{ color: 'var(--text-muted)' }} />
                <p style={{ fontWeight: '500' }}>Search your Knowledge Base</p>
                <p style={{ fontSize: '0.85rem', marginTop: '-0.5rem' }}>
                  Upload a PDF document on the left, then search above to find semantically relevant contents.
                </p>
              </div>
            ) : searchResults.length === 0 ? (
              <div className="results-status-info">
                <AlertCircle size={48} style={{ color: 'var(--text-muted)' }} />
                <p style={{ fontWeight: '500' }}>No matching results found</p>
                <p style={{ fontSize: '0.85rem', marginTop: '-0.5rem' }}>
                  Try uploading another document or tweaking your search terms.
                </p>
              </div>
            ) : (
              searchResults.map((result, idx) => {
                const source = result.metadata?.source || 'Unknown Source';
                const page = result.metadata?.page || '?';
                
                // Let's perform a very basic word highlight for visual look
                return (
                  <div key={idx} className="result-card">
                    <div className="result-header">
                      <div className="result-meta-left">
                        <span className="result-tag source">{source}</span>
                        <span className="result-tag page">Page {page}</span>
                      </div>
                      {result.score > 0 && (
                        <div className="result-score">
                          Match: {Math.round(result.score * 100)}%
                        </div>
                      )}
                    </div>
                    <div className="result-content">
                      {result.content}
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </section>

      </main>
    </div>
  );
}

export default App;
