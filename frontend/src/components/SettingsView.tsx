import React, { useState, useEffect } from 'react';
import { GetOllamaPath, UpdateOllamaPath, SelectFolder, DownloadModel, RemoveModel } from '../../wailsjs/go/main/App';

interface SettingsViewProps {
  availableModels: string[];
  fetchModels: () => Promise<void>;
  popularModels: string[];
  downloadProgress: Record<string, string>;
  setDownloadProgress: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  isDownloading: Record<string, boolean>;
  setIsDownloading: React.Dispatch<React.SetStateAction<Record<string, boolean>>>;
}

const SettingsView: React.FC<SettingsViewProps> = ({ 
  availableModels, 
  fetchModels, 
  popularModels,
  downloadProgress,
  setDownloadProgress,
  isDownloading,
  setIsDownloading
}) => {
  const [ollamaPath, setOllamaPath] = useState('');
  const [isModelsExpanded, setIsModelsExpanded] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modelToRemove, setModelToRemove] = useState('');

  useEffect(() => {
    const fetchPath = async () => {
      try {
        const path = await GetOllamaPath();
        setOllamaPath(path);
      } catch (err) {
        console.error(err);
      }
    };
    fetchPath();
  }, []);

  const handleBrowseFolder = async () => {
    try {
      const selected = await SelectFolder();
      if (selected) {
        setOllamaPath(selected);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleSavePath = async () => {
    try {
      const success = await UpdateOllamaPath(ollamaPath);
      if (success) {
        alert('Settings saved. Restarting Ollama engine...');
        setTimeout(() => {
          fetchModels();
        }, 3000);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const startDownload = async (modelName: string) => {
    setIsDownloading(prev => ({ ...prev, [modelName]: true }));
    setDownloadProgress(prev => ({ ...prev, [modelName]: 'Starting...' }));
    
    try {
      const result = await DownloadModel(modelName);
      if (result === "Success") {
        fetchModels();
      } else {
        alert(`Failed to download ${modelName}.`);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsDownloading(prev => ({ ...prev, [modelName]: false }));
    }
  };

  const handleRemoveModel = (modelName: string) => {
    const exactName = availableModels.find(m => m === modelName || m.startsWith(modelName + ':')) || modelName;
    setModelToRemove(exactName);
    setIsModalOpen(true);
  };

  const confirmRemoveModel = async () => {
    setIsModalOpen(false);
    const result = await RemoveModel(modelToRemove);
    if (result === "Success") {
      fetchModels();
    } else {
      alert(`Failed to remove model: ${result}`);
    }
  };

  return (
    <div className="placeholder-view" style={{ width: '100%', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Settings</h2>
      <div className="settings-separator"></div>
      
      <div style={{ width: '100%', marginBottom: '40px' }}>
        <label style={{ display: 'block', marginBottom: '8px', fontWeight: '600', textAlign: 'left' }}>Ollama Models Path</label>
        <div className="settings-path-container">
          <input 
            type="text" 
            value={ollamaPath} 
            onChange={(e) => setOllamaPath(e.target.value)}
            className="settings-input-field" 
          />
          <button onClick={handleBrowseFolder} className="settings-browse-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
            </svg>
          </button>
        </div>
        <button onClick={handleSavePath} className="chat-submit-btn" style={{ borderRadius: '8px', width: 'auto', height: 'auto', padding: '12px 24px' }}>
          Save Settings
        </button>
      </div>

      <div style={{ width: '100%' }}>
        <h3 style={{ fontSize: '18px', color: '#2b2722', marginBottom: '16px', textAlign: 'left' }}>Model Management</h3>
        
        <div className="models-accordion">
          <div 
            className="accordion-header"
            onClick={() => setIsModelsExpanded(!isModelsExpanded)}
          >
            <span>Popular Models</span>
            <span>{isModelsExpanded ? '▲' : '▼'}</span>
          </div>
          
          {isModelsExpanded && (
            <div className="accordion-content">
              {popularModels.map((modelName) => {
                const isInstalled = availableModels.includes(modelName) || availableModels.some(m => m.startsWith(modelName + ':'));
                const downloading = isDownloading[modelName];
                const progress = downloadProgress[modelName];

                return (
                  <div key={modelName} className="model-list-item">
                    <div className="model-info">
                      <span className="model-name">{modelName}</span>
                      {isInstalled ? (
                        <span className="status-badge status-installed">Installed</span>
                      ) : (
                        <span className="status-badge status-not-installed">Not Installed</span>
                      )}
                    </div>
                    
                    <div className="action-area">
                      {downloading ? (
                        progress && progress.includes('%') ? (
                          <div className="progress-container">
                            <div className="progress-bar">
                              <div className="progress-fill" style={{ width: progress }}></div>
                            </div>
                            <span className="progress-percentage">{progress}</span>
                          </div>
                        ) : (
                          <button className="download-btn starting" disabled>
                            Starting...
                          </button>
                        )
                      ) : isInstalled ? (
                        <button 
                          onClick={() => handleRemoveModel(modelName)}
                          className="remove-btn"
                        >
                          Remove
                        </button>
                      ) : (
                        <button 
                          onClick={() => startDownload(modelName)}
                          className="download-btn"
                        >
                          Download
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {isModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <h3>Remove Model</h3>
            <p>Are you sure you want to remove <strong>{modelToRemove}</strong>?</p>
            <div className="modal-actions">
              <button className="modal-btn-cancel" onClick={() => setIsModalOpen(false)}>
                Cancel
              </button>
              <button className="modal-btn-confirm" onClick={confirmRemoveModel}>
                Remove
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SettingsView;