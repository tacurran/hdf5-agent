import React, { useState, useEffect, useCallback } from 'react';
import './App.css';

// Icons as inline SVG components
const Icons = {
  File: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
      <polyline points="13 2 13 9 20 9"></polyline>
    </svg>
  ),
  Folder: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
    </svg>
  ),
  Database: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
      <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
      <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
    </svg>
  ),
  ChevronRight: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <polyline points="9 18 15 12 9 6"></polyline>
    </svg>
  ),
  ChevronDown: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <polyline points="6 9 12 15 18 9"></polyline>
    </svg>
  ),
  Edit: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
    </svg>
  ),
  Download: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
      <polyline points="7 10 12 15 17 10"></polyline>
      <line x1="12" y1="15" x2="12" y2="3"></line>
    </svg>
  ),
  Plus: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <line x1="12" y1="5" x2="12" y2="19"></line>
      <line x1="5" y1="12" x2="19" y2="12"></line>
    </svg>
  ),
  Trash: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <polyline points="3 6 5 6 21 6"></polyline>
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
      <line x1="10" y1="11" x2="10" y2="17"></line>
      <line x1="14" y1="11" x2="14" y2="17"></line>
    </svg>
  ),
};

const TreeNode = ({ node, path = "", onSelect, selectedPath, level = 0 }) => {
  const [expanded, setExpanded] = useState(false);
  const isGroup = node.type === 'group';
  const isFile = node.type === 'file';
  const isDataset = node.type === 'dataset';
  const currentPath = path ? `${path}/${node.name}` : node.name;
  const isSelected = selectedPath === currentPath;

  const handleClick = () => {
    if (!isGroup) {
      onSelect(currentPath, node);
    } else {
      setExpanded(!expanded);
    }
  };

  return (
    <div key={currentPath} className="tree-item">
      <div
        className={`tree-node ${isSelected ? 'selected' : ''}`}
        onClick={handleClick}
        style={{ paddingLeft: `${level * 16}px` }}
      >
        {isGroup && (
          <span className="tree-toggle">
            {expanded ? <Icons.ChevronDown /> : <Icons.ChevronRight />}
          </span>
        )}
        {!isGroup && <span className="tree-toggle-placeholder" />}

        {isFile && <Icons.File className="tree-icon" />}
        {isGroup && <Icons.Folder className="tree-icon" />}
        {isDataset && <Icons.Database className="tree-icon" />}

        <span className="tree-label">{node.name}</span>
        {isDataset && node.shape && (
          <span className="tree-meta">{node.shape.join('×')}</span>
        )}
      </div>

      {isGroup && expanded && node.children && (
        <div className="tree-children">
          {node.children.map((child, idx) => (
            <TreeNode
              key={idx}
              node={child}
              path={currentPath}
              onSelect={onSelect}
              selectedPath={selectedPath}
              level={level + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
};

const DataViewer = ({ data, onUpdate }) => {
  const [editingIndex, setEditingIndex] = useState(null);
  const [editValue, setEditValue] = useState('');

  if (!data) return null;

  const renderData = (value) => {
    if (Array.isArray(value)) {
      return (
        <div className="data-array">
          {value.slice(0, 100).map((item, idx) => (
            <div
              key={idx}
              className={`data-cell ${editingIndex === idx ? 'editing' : ''}`}
              onClick={() => {
                setEditingIndex(idx);
                setEditValue(String(item));
              }}
            >
              {editingIndex === idx ? (
                <input
                  type="text"
                  value={editValue}
                  onChange={(e) => setEditValue(e.target.value)}
                  onBlur={() => {
                    if (editValue !== String(value[idx])) {
                      onUpdate(idx, editValue);
                    }
                    setEditingIndex(null);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      onUpdate(idx, editValue);
                      setEditingIndex(null);
                    }
                  }}
                  autoFocus
                />
              ) : (
                <span>{typeof item === 'number' ? item.toFixed(4) : item}</span>
              )}
            </div>
          ))}
          {value.length > 100 && <div className="data-truncated">... and {value.length - 100} more</div>}
        </div>
      );
    }

    if (typeof value === 'object' && value !== null) {
      return <pre className="data-json">{JSON.stringify(value, null, 2)}</pre>;
    }

    return <div className="data-value">{String(value)}</div>;
  };

  return <div className="data-viewer">{renderData(data)}</div>;
};

export default function App() {
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [fileStructure, setFileStructure] = useState(null);
  const [selectedDataset, setSelectedDataset] = useState(null);
  const [datasetData, setDatasetData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showNewFileModal, setShowNewFileModal] = useState(false);
  const [newFilename, setNewFilename] = useState('');
  const [selectedPath, setSelectedPath] = useState(null);

  const API_BASE = 'http://localhost:8080/api';

  const loadFiles = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE}/files`);
      const data = await response.json();
      setFiles(data || []);
      setError('');
    } catch (err) {
      setError('Failed to load files: ' + err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadFileStructure = useCallback(async (filePath) => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE}/file/structure?path=${encodeURIComponent(filePath)}`);
      const data = await response.json();
      setFileStructure(data);
      setSelectedDataset(null);
      setDatasetData(null);
      setError('');
    } catch (err) {
      setError('Failed to load file structure: ' + err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadDataset = useCallback(async (filePath, datasetPath) => {
    try {
      setLoading(true);
      const response = await fetch(
        `${API_BASE}/dataset/read?file=${encodeURIComponent(filePath)}&path=${encodeURIComponent(datasetPath)}`
      );
      const data = await response.json();
      setDatasetData(data);
      setError('');
    } catch (err) {
      setError('Failed to load dataset: ' + err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleSelectFile = async (file) => {
    setSelectedFile(file);
    setSelectedPath(file.name);
    await loadFileStructure(file.path);
  };

  const handleSelectDataset = async (path, node) => {
    setSelectedPath(path);
    setSelectedDataset({ path, ...node });
    await loadDataset(selectedFile.path, path);
  };

  const handleUpdateData = async (index, newValue) => {
    try {
      const response = await fetch(`${API_BASE}/dataset/update?file=${encodeURIComponent(selectedFile.path)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: selectedDataset.path,
          data: [newValue],
          indices: [index],
        }),
      });

      if (response.ok) {
        setError('');
        // Reload dataset to show updated value
        await loadDataset(selectedFile.path, selectedDataset.path);
      } else {
        setError('Failed to update data');
      }
    } catch (err) {
      setError('Update failed: ' + err.message);
    }
  };

  const handleCreateFile = async () => {
    if (!newFilename.trim()) return;

    try {
      const response = await fetch(`${API_BASE}/file/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filename: newFilename }),
      });

      if (response.ok) {
        setNewFilename('');
        setShowNewFileModal(false);
        await loadFiles();
      } else {
        setError('Failed to create file');
      }
    } catch (err) {
      setError('Create file failed: ' + err.message);
    }
  };

  const handleDeleteFile = async (filePath) => {
    if (!confirm('Delete file? This cannot be undone.')) return;

    try {
      const response = await fetch(`${API_BASE}/file/delete?path=${encodeURIComponent(filePath)}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setSelectedFile(null);
        setFileStructure(null);
        await loadFiles();
      } else {
        setError('Failed to delete file');
      }
    } catch (err) {
      setError('Delete failed: ' + err.message);
    }
  };

  useEffect(() => {
    loadFiles();
  }, [loadFiles]);

  return (
    <div className="app">
      {/* Header */}
      <header className="app-header">
        <div className="header-content">
          <div className="header-title">
            <Icons.Database />
            <h1>HDF5 Agent</h1>
            <span className="header-subtitle">Data Manipulation Interface</span>
          </div>
          <button className="btn btn-primary" onClick={() => setShowNewFileModal(true)}>
            <Icons.Plus /> New File
          </button>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="error-banner">
          <span>{error}</span>
          <button onClick={() => setError('')}>×</button>
        </div>
      )}

      <div className="app-container">
        {/* Sidebar - File Browser */}
        <aside className="sidebar">
          <div className="sidebar-header">
            <h2>Files</h2>
            {loading && <div className="spinner" />}
          </div>

          <div className="file-list">
            {files && files.length > 0 ? (
              files.map((file) => (
                <div
                  key={file.path}
                  className={`file-item ${selectedFile?.path === file.path ? 'active' : ''}`}
                >
                  <div
                    className="file-item-header"
                    onClick={() => handleSelectFile(file)}
                  >
                    <Icons.File />
                    <div className="file-item-info">
                      <div className="file-name">{file.name}</div>
                      <div className="file-size">{(parseInt(file.size) / 1024).toFixed(1)} KB</div>
                    </div>
                  </div>
                  {selectedFile?.path === file.path && (
                    <button
                      className="btn-delete"
                      onClick={() => handleDeleteFile(file.path)}
                      title="Delete file"
                    >
                      <Icons.Trash />
                    </button>
                  )}
                </div>
              ))
            ) : (
              <div className="empty-state">No HDF5 files found</div>
            )}
          </div>
        </aside>

        {/* Main Content */}
        <main className="main-content">
          <div className="content-layout">
            {/* Tree View */}
            {selectedFile && fileStructure ? (
              <div className="tree-pane">
                <div className="pane-header">
                  <h2>Structure</h2>
                  <span className="file-badge">{selectedFile.name}</span>
                </div>
                <div className="tree-view">
                  {fileStructure.children && fileStructure.children.length > 0 ? (
                    fileStructure.children.map((child, idx) => (
                      <TreeNode
                        key={idx}
                        node={child}
                        onSelect={handleSelectDataset}
                        selectedPath={selectedPath}
                      />
                    ))
                  ) : (
                    <div className="empty-state">No datasets in this file</div>
                  )}
                </div>
              </div>) : (
              <div className="tree-pane empty">
                <div className="empty-state">
                  <Icons.Folder />
                  <p>Select a file to view structure</p>
                </div>
              </div>
            )}

            {/* Data View */}
            {selectedDataset ? (
              <div className="data-pane">
                <div className="pane-header">
                  <div className="header-info">
                    <h2>{selectedDataset.name}</h2>
                    <div className="data-meta">
                      <span className="meta-item">Shape: {selectedDataset.shape?.join(' × ')}</span>
                      <span className="meta-item">Type: {selectedDataset.datatype}</span>
                      <span className="meta-item">Size: {selectedDataset.size}</span>
                    </div>
                  </div>
                  <button className="btn btn-secondary">
                    <Icons.Download /> Export
                  </button>
                </div>

                {loading ? (
                  <div className="loading-state">
                    <div className="spinner" />
                    <p>Loading data...</p>
                  </div>
                ) : datasetData ? (
                  <DataViewer
                    data={datasetData.data}
                    onUpdate={handleUpdateData}
                  />
                ) : (
                  <div className="empty-state">No data</div>
                )}
              </div>
            ) : (
              <div className="data-pane empty">
                <div className="empty-state">
                  <Icons.Database />
                  <p>Select a dataset to view data</p>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>

      {/* New File Modal */}
      {showNewFileModal && (
        <div className="modal-overlay" onClick={() => setShowNewFileModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Create New File</h2>
              <button
                className="btn-close"
                onClick={() => setShowNewFileModal(false)}
              >
                ×
              </button>
            </div>
            <div className="modal-body">
              <input
                type="text"
                placeholder="filename.h5"
                value={newFilename}
                onChange={(e) => setNewFilename(e.target.value)}
                className="input-text"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleCreateFile();
                }}
                autoFocus
              />
            </div>
            <div className="modal-footer">
              <button
                className="btn btn-secondary"
                onClick={() => setShowNewFileModal(false)}
              >
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleCreateFile}>
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
